package emulator

import (
	"fmt"
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
		return nil
	case <-timeout.C:
		timeout.Stop()
		return fmt.Errorf("joypad timedout after 1s to finish")
	}
}

func (j *Joypad) init() {

	// no button selected, all keys released
	j.c.memory.Write(PORT_JOYPAD, 0x3F)

	// create IO ticker
	ticker := time.NewTicker(50 * time.Millisecond)

	go func(c *Cpu, stopChan, doneChan chan bool) {

		for {
			select {
			case <-stopChan:
				ticker.Stop()
				doneChan <- true
				return
			case <-ticker.C:

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
						case rl.KeyQ:
							// start
							jp &= 0xB
						case rl.KeyW:
							// select
							jp &= 0x7
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
			}
		}

	}(j.c, j.stopChan, j.doneChan)
}
