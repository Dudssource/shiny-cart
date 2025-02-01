package emulator

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameBoy struct {
	// components
	c      *Cpu
	joypad *Joypad
	timer  *Timer
	video  *Video
}

func NewGameBoy(debug, step, silent bool, breakPoints string) *GameBoy {
	if step {
		debug = true
	}
	c := &Cpu{
		step:        step,
		silent:      silent,
		breakPoints: strings.TrimSpace(breakPoints),
		debug:       debug,
		memory:      NewMemory(),
	}

	return &GameBoy{
		c:      c,
		joypad: NewJoypad(c),
		timer:  NewTimer(c),
		video:  &Video{},
	}
}

func (g *GameBoy) Load(romFile string) error {

	f, err := os.Open(romFile)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	if len(g.c.memory.mem) == 0 {
		g.c.memory.mem = make(memoryArea, 65536)
	}

	if len(g.c.memory.rom) == 0 {
		g.c.memory.rom = make(memoryArea, stat.Size())
	}

	address := 0x0
	for {

		b := make([]byte, 1)
		if _, err := f.Read(b); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		for _, value := range b {
			if address < int(VRAM_START) {
				g.c.memory.mem[address] = value
			}
			g.c.memory.rom[address] = value
			address++
		}
	}
}

// Game Loop
func (g *GameBoy) Loop(interval time.Duration) error {

	// init emulator
	if err := g.init(); err != nil {
		return err
	}

	stop := make(chan byte, 1)
	fps := make(chan byte, 1)

	go func(g *GameBoy, fps, stop chan byte) {

		defer func() {
			stop <- 0x0
		}()

		totalCycles := 0

		for {
			select {
			case <-stop:
				return
			case <-fps:
				cycle := 0

				for cycle < 17476 {

					// broadcast machine cycle
					g.broadcast(totalCycles)

					if g.c.stopped {
						return
					}

					// overflow internal m-cycle counter, reset
					if totalCycles+1 > math.MaxInt32 {
						totalCycles = 0
					} else {
						totalCycles++
					}

					cycle++

					if rl.IsKeyPressed(rl.KeyP) {
						g.c.step = true
					}
				}
			}
		}
	}(g, fps, stop)

	// block
	for !rl.WindowShouldClose() && !g.c.stopped {

		if len(fps) == 0 {
			fps <- 0x0
		}

		// emulate raylib event loop
		g.video.Draw([]byte{}, 0, 0, false)
	}

	if !g.c.stopped {
		log.Println("STOPPING")
		stop <- 0x0
		timeout := time.After(time.Second)
		select {
		case <-stop:
			log.Println("Cpu m-clycle emulator stopped")
		case <-timeout:
			return fmt.Errorf("timedout after waiting %s waiting for cpu m-cycle emulator to stop", time.Second.String())
		}
	}

	// close window
	rl.CloseWindow()

	// force stop the emulator
	return nil
}

func (g *GameBoy) init() error {

	if err := g.c.memory.init(); err != nil {
		return err
	}

	// init handlers
	g.video.init(512, 256)
	if err := g.c.init(); err != nil {
		return err
	}
	g.joypad.init()
	g.timer.init()

	return nil
}

func (g *GameBoy) broadcast(cycle int) {
	g.joypad.sync(cycle)
	g.timer.sync(cycle)
	g.c.sync(cycle)
}
