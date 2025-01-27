package emulator

import (
	"io"
	"log"
	"math"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameBoy struct {
	// components
	c      *Cpu
	joypad *Joypad
	timer  *Timer
	video  *Video

	// Channels
	joypadChan chan int
	cpuChan    chan int
	timerChan  chan int

	cycle int
}

func NewGameBoy(debug, step bool) *GameBoy {
	if step {
		debug = true
	}
	c := &Cpu{
		step:   step,
		debug:  debug,
		memory: NewMemory(),
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

	start := 0x0
	for {

		b := make([]byte, 1)
		if _, err := f.Read(b); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		for _, b1 := range b {
			g.c.memory.Write(Word(start), b1)
			start++
		}
	}
}

// Game Loop
func (g *GameBoy) Loop(interval time.Duration) error {

	// init emulator
	if err := g.init(); err != nil {
		return err
	}

	// machine cycle channel
	// one cicle = 1us
	mCycle := time.NewTicker(interval)

	stopChan := make(chan bool, 1)

	go func(g *GameBoy, stopChan <-chan bool) {
		for {
			select {
			case <-stopChan:
				return

			case <-mCycle.C:

				// broadcast machine cycle
				g.broadcast()
			}
		}
	}(g, stopChan)

	// block
	for !rl.WindowShouldClose() && !g.c.stopped {
		// emulate raylib event loop
		g.video.Draw([]byte{}, 0, 0, false)
	}

	log.Println("STOPPING")

	// stop mCycle
	mCycle.Stop()

	// close window
	rl.CloseWindow()

	// force stop the emulator
	return g.stop()
}

func (g *GameBoy) init() error {

	if err := g.c.memory.init(); err != nil {
		return err
	}

	// init channels
	g.cpuChan = make(chan int, 1)
	g.joypadChan = make(chan int, 1)
	g.timerChan = make(chan int, 1)

	// init handlers
	g.video.init(512, 256)
	g.c.init(g.cpuChan)
	g.joypad.init(g.joypadChan)
	g.timer.init(g.timerChan)

	return nil
}

func (g *GameBoy) stop() error {
	// close joypad
	if err := g.joypad.stop(); err != nil {
		return err
	}
	// close timer
	if err := g.timer.stop(); err != nil {
		return err
	}
	if !g.c.stopped {
		// close cpu
		if err := g.c.stop(); err != nil {
			return err
		}
	}
	// close channels
	close(g.joypadChan)
	close(g.timerChan)
	close(g.cpuChan)

	// no errors
	return nil
}

func (g *GameBoy) broadcast() {

	// overflow internal m-cycle counter, reset
	if g.cycle+1 > math.MaxInt32 {
		g.cycle = 0
	} else {
		g.cycle++
	}

	g.joypadChan <- g.cycle
	g.timerChan <- g.cycle
	g.cpuChan <- g.cycle
}
