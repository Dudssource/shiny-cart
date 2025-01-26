package emulator

import (
	"fmt"
	"log"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	PORT_JOYPAD = Word(0xFF00)
)

type Joypad struct {
	c        *Cpu
	mu       *sync.Mutex
	stopChan chan bool
	doneChan chan bool
}

func NewJoypad(c *Cpu) *Joypad {
	return &Joypad{
		c:        c,
		mu:       &sync.Mutex{},
		stopChan: make(chan bool),
		doneChan: make(chan bool),
	}
}

func (j *Joypad) stop() error {
	// ask to stop
	j.stopChan <- true
	// specifies a timeout
	timeout := time.NewTimer(time.Second)

	select {
	// wait until Joypad stops
	case <-j.doneChan:
		// close channels
		close(j.stopChan)
		close(j.doneChan)
		return nil
	case <-timeout.C:
		timeout.Stop()
		return fmt.Errorf("joypad timedout after 1s to finish")
	}
}

func (j *Joypad) init(mCycle <-chan int) {

	// no button selected, all keys released
	j.c.memory.Write(PORT_JOYPAD, 0x3F)

	go func(c *Cpu, mCyche <-chan int, stopChan, doneChan chan bool) {

		defer func() {
			doneChan <- true
		}()

		for {
			select {
			case <-stopChan:
				log.Println("JOYPAD STOPPED")
				return
			case <-mCycle:

				// read joypad register
				jp := c.memory.Read(PORT_JOYPAD)

				// no buttons selected
				if jp == 0x30 {
					// all keys released
					c.memory.Write(PORT_JOYPAD, 0x3F)
					// proceed
					continue
				}

				// unset the lowest nibbles (for game boy, 0 means key pressed)
				jp |= 0xF

				// check if select buttons
				if jp&0x20 == 0x0 {
					for {
						key := rl.GetKeyPressed()
						if key == 0 {
							break
						}

						switch key {
						case rl.KeyW:
							// select
							jp &= 0x7
						case rl.KeyQ:
							// start
							jp &= 0xB
						case rl.KeyA:
							// B
							jp &= 0xD
						case rl.KeyS:
							// A
							jp &= 0xE
						}
					}

					// directional
				} else if jp&0x10 == 0x0 {
					for {
						key := rl.GetKeyPressed()
						if key == 0 {
							break
						}

						switch key {
						case rl.KeyDown:
							// down
							jp &= 0x7
						case rl.KeyUp:
							// up
							jp &= 0xB
						case rl.KeyLeft:
							// left
							jp &= 0xD
						case rl.KeyRight:
							// right
							jp &= 0xE
						}
					}
				}

				if jp&0xF != 0xF {
					// request interrupt
					iflag := c.memory.Read(INTERRUPT_FLAG)
					iflag |= 0x10
					c.memory.Write(INTERRUPT_FLAG, iflag)
				}
			}
		}

	}(j.c, mCycle, j.stopChan, j.doneChan)
}
