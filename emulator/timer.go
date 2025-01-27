package emulator

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	PORT_DIV  = Word(0xFF04)
	PORT_TIMA = Word(0xFF05)
	PORT_TMA  = Word(0xFF06)
	PORT_TAC  = Word(0xFF07)
)

var (
	// https://gbdev.io/pandocs/Timer_and_Divider_Registers.html#ff07--tac-timer-control
	timaClock = map[uint8]int{
		0x0: 256,
		0x1: 4,
		0x2: 16,
		0x3: 64,
	}
)

type Timer struct {
	c        *Cpu
	mu       *sync.Mutex
	stopChan chan bool
	doneChan chan bool
}

func NewTimer(c *Cpu) *Timer {
	return &Timer{
		c:        c,
		mu:       &sync.Mutex{},
		stopChan: make(chan bool),
		doneChan: make(chan bool),
	}
}

func (t *Timer) stop() error {
	// ask to stop
	t.stopChan <- true
	// specifies a timeout
	timeout := time.NewTimer(time.Second)

	select {
	// wait until timer stops
	case <-t.doneChan:
		// close channels
		close(t.stopChan)
		close(t.doneChan)
		return nil
	case <-timeout.C:
		timeout.Stop()
		return fmt.Errorf("timer timedout after 1s to finish")
	}
}

func (t *Timer) init(mCycle <-chan int) {

	go func(c *Cpu, mCycle <-chan int, stopChan, doneChan chan bool) {

		defer func() {
			doneChan <- true
		}()

		lastDivValue := 0

		for {
			select {
			case cycle := <-mCycle:

				/* DIV REGISTER */
				currentDivValue := int(c.memory.Read(PORT_DIV))
				// some value was written to DIV or we had an overflow
				if lastDivValue > 0 && lastDivValue != currentDivValue {
					lastDivValue = 0
					// Reset DIV
					c.memory.Write(PORT_DIV, 0)
				} else if cycle%64 == 0 { // 16384 Hz
					// overflow
					if currentDivValue+1 > 0xFF {
						// Reset DIV
						c.memory.Write(PORT_DIV, 0)
					} else {
						currentDivValue++
						// increment DIV
						c.memory.Write(PORT_DIV, uint8(currentDivValue))
						lastDivValue = currentDivValue
					}
				}

				/* TIMA REGISTER */
				tac := c.memory.Read(PORT_TAC)
				timaEnabled := (tac & 0x4) > 0

				if timaEnabled {
					timaFrequency := tac & 0x3
					if cycle%timaClock[timaFrequency] == 0 {
						currentTimaValue := int(c.memory.Read(PORT_TIMA))
						// overflow
						if currentTimaValue+1 > 0xFF {
							tma := c.memory.Read(PORT_TMA)
							iflag := c.memory.Read(INTERRUPT_FLAG) | 0x4
							log.Printf("TIMA=%d, CYCLE %d, FREQ=%d, TMA=%d, IFLAG=%.8b\n", currentTimaValue, cycle, timaClock[timaFrequency], tma, iflag)
							// set TIMA as TMA (Modulo)
							c.memory.Write(PORT_TIMA, tma)
							// request TIMER interruption
							c.memory.Write(INTERRUPT_FLAG, iflag)
						} else {
							// increment TIMA
							c.memory.Write(PORT_TIMA, uint8(currentTimaValue+1))
						}
					}
				}

			case <-stopChan:
				log.Println("TIMER STOPPED")
				return
			}
		}

	}(t.c, mCycle, t.stopChan, t.doneChan)
}
