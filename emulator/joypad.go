package emulator

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	PORT_JOYPAD = Word(0xFF00)
)

type Joypad struct {
	c *Cpu
}

func NewJoypad(c *Cpu) *Joypad {
	return &Joypad{
		c: c,
	}
}

func (j *Joypad) sync(_ int) {

	// read joypad register
	jp := j.c.memory.Read(PORT_JOYPAD)

	// no buttons selected
	if jp == 0x30 {
		// all keys released
		j.c.memory.Write(PORT_JOYPAD, 0x3F)
		// proceed
		return
	}

	// unset the lowest nibbles (for game boy, 0 means key pressed)
	jp = (jp & 0xF0) | 0xF

	// check if select buttons
	if jp&0x20 > 0x0 {
		for {
			key := rl.GetKeyPressed()
			if key == 0 {
				break
			}

			switch key {
			case rl.KeyW:
				// select
				jp &= 0xF7
			case rl.KeyQ:
				// start
				jp &= 0xFB
			case rl.KeyA:
				// B
				jp &= 0xFD
			case rl.KeyS:
				// A
				jp &= 0xFE
			}
		}

		// directional
	} else if jp&0x10 > 0x0 {
		for {
			key := rl.GetKeyPressed()
			if key == 0 {
				break
			}

			switch key {
			case rl.KeyDown:
				// down
				jp &= 0xF7
			case rl.KeyUp:
				// up
				jp &= 0xFB
			case rl.KeyLeft:
				// left
				jp &= 0xFD
			case rl.KeyRight:
				// right
				jp &= 0xFE
			}
		}
	}

	if jp&0xF != 0xF {
		// request interrupt
		iflag := j.c.memory.Read(INTERRUPT_FLAG)
		iflag |= 0x10
		j.c.memory.Write(INTERRUPT_FLAG, iflag)
	}
}

func (j *Joypad) init() {

	// no button selected, all keys released
	j.c.memory.Write(PORT_JOYPAD, 0x3F)
}
