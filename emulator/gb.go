package emulator

import (
	"fmt"
	"os"
)

type GameBoy struct {
	c      *Cpu
	joypad *Joypad
}

// Game Loop
func (g *GameBoy) Loop() {
	// stop joypad
	if err := g.joypad.stop(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
}
