package emulator

import (
	"fmt"
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
	sound  *Sound
}

func NewGameBoy(debug, step, silent, profiling bool, breakPoints string, palette, channels int) *GameBoy {
	if step {
		debug = true
	}

	mem := make(memoryArea, 65536)
	sound := NewSound(mem, channels)

	c := &Cpu{
		step:        step,
		silent:      silent,
		breakPoints: strings.TrimSpace(breakPoints),
		debug:       debug,
		profiling:   profiling,
		memory:      NewMemory(sound, mem),
	}

	return &GameBoy{
		c:      c,
		timer:  NewTimer(c),
		joypad: NewJoypad(c.memory),
		sound:  sound,
		video: &Video{
			mem:     c.memory,
			mode:    2,
			palette: palette,
		},
	}
}

func (g *GameBoy) Load(romFile string) error {

	f1, err := os.ReadFile(romFile)
	if err != nil {
		return err
	}

	if len(g.c.memory.mem) == 0 {
		g.c.memory.mem = make(memoryArea, 65536)
	}

	if len(g.c.memory.rom) == 0 {
		g.c.memory.rom = make(memoryArea, len(f1))
	}

	for address, value := range f1 {
		if address < int(VRAM_START) {
			g.c.memory.mem[address] = value
		}
		g.c.memory.rom[address] = value
	}

	return nil
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

		tCycles := 0
		mCycles := 0

		start := time.Now()
		cyclesPerSecond := 0
		ticks := 0

		for {
			select {
			case <-stop:
				return
			case <-fps:

				// 4.194304 MHz / 60 FPS
				for range 69905 {

					if tCycles%4 == 0 {

						// broadcast machine cycle
						g.broadcast(mCycles)

						// 4Mihz (t-cycles) = 1 Mihz (m-cycles) == 1ms
						if mCycles%1048 == 0 {
							ticks++
							// used for tshoot and profiling
							if ticks == 1000 {
								if g.c.profiling {
									log.Println("Tick RTC after 1s")
								}
								ticks = 0
							}

							if g.c.memory.mbc.initialized() {
								// RTC tick (if supported by cartridge)
								g.c.memory.mbc.controller.Tick()
							}
						}

						if g.c.stopped {
							return
						}

						// overflow internal m-cycle counter, reset
						if mCycles == math.MaxInt32 {
							mCycles = 0
						} else {
							mCycles++
						}

						if time.Since(start).Seconds() >= 1 {
							if g.c.profiling {
								log.Printf("M-cycles per second %d\n", cyclesPerSecond)
							}
							cyclesPerSecond = 0
							start = time.Now()
						}
						cyclesPerSecond++

						if rl.IsKeyPressed(rl.KeyP) {
							g.c.step = true
						}
					}

					// timer v2
					g.timer.sync2(tCycles)

					// every T-cycle
					g.sound.sync(tCycles)

					if tCycles == math.MaxInt32 {
						tCycles = 0
					} else {
						tCycles++
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
		g.video.draw()
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

	// stop sound
	g.sound.stop()

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
	g.video.init(640, 576)
	if err := g.c.init(); err != nil {
		return err
	}
	g.joypad.init()
	g.timer.init()
	return g.sound.init()
}

func (g *GameBoy) broadcast(cycle int) {
	g.joypad.sync(cycle)
	g.c.sync(cycle)
	//g.timer.sync(cycle)
	g.video.scan(g.c)
}
