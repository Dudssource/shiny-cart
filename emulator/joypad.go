package emulator

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	PORT_JOYPAD = Word(0xFF00)
)

type Joypad struct {
	memory *Memory
}

func NewJoypad(memory *Memory) *Joypad {
	return &Joypad{
		memory: memory,
	}
}

func (j *Joypad) sync(_ int) {
	jp := j.memory.mem[PORT_JOYPAD]

	if j.memory.joypad&0x80 == 0 && rl.IsKeyReleased(rl.KeyQ) {
		j.memory.joypad |= 0x80
	}

	if j.memory.joypad&0x40 == 0 && rl.IsKeyReleased(rl.KeyW) {
		j.memory.joypad |= 0x40
	}

	if j.memory.joypad&0x20 == 0 && rl.IsKeyReleased(rl.KeyS) {
		j.memory.joypad |= 0x20
	}

	if j.memory.joypad&0x10 == 0 && rl.IsKeyReleased(rl.KeyA) {
		j.memory.joypad |= 0x10
	}

	if j.memory.joypad&0x8 == 0 && rl.IsKeyReleased(rl.KeyDown) {
		j.memory.joypad |= 0x8
	}

	if j.memory.joypad&0x4 == 0 && rl.IsKeyReleased(rl.KeyUp) {
		j.memory.joypad |= 0x4
	}

	if j.memory.joypad&0x2 == 0 && rl.IsKeyReleased(rl.KeyLeft) {
		j.memory.joypad |= 0x2
	}

	if j.memory.joypad&0x1 == 0 && rl.IsKeyReleased(rl.KeyRight) {
		j.memory.joypad |= 0x1
	}

	var pressed uint8

	switch rl.GetKeyPressed() {
	case rl.KeyQ:
		// start
		j.memory.joypad &= 0x7F
		pressed |= 0x1 // button
	case rl.KeyW:
		// select
		j.memory.joypad &= 0xBF
		pressed |= 0x1 // button
	case rl.KeyA:
		// A
		j.memory.joypad &= 0xEF
		pressed |= 0x1 // button
	case rl.KeyS:
		// B
		j.memory.joypad &= 0xDF
		pressed |= 0x1 // button
	case rl.KeyDown:
		// down
		j.memory.joypad &= 0xF7
		pressed |= 0x2 // directional
	case rl.KeyUp:
		// up
		j.memory.joypad &= 0xFB
		pressed |= 0x2 // directional
	case rl.KeyLeft:
		// left
		j.memory.joypad &= 0xFD
		pressed |= 0x2 // directional
	case rl.KeyRight:
		// right
		j.memory.joypad &= 0xFE
		pressed |= 0x2 // directional
	}

	// button pressed OR directional
	if (jp&0x20 == 0x0 && (pressed&0x1) > 0) || (jp&0x10 == 0x0 && (pressed&0x2) > 0) {
		// request interrupt
		iflag := j.memory.Read(INTERRUPT_FLAG)
		iflag |= 0x10
		j.memory.Write(INTERRUPT_FLAG, iflag)
	}
}

func (j *Joypad) init() {}
