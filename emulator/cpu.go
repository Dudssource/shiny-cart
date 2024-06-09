package emulator

import (
	"fmt"
	"time"
)

type Mode string

const (
	DMG  Mode = "DMG"
	MGB  Mode = "MGB"
	SGB  Mode = "SGB"
	SGB2 Mode = "SGB2"
	CGB  Mode = "CGB"
	AGB  Mode = "AGB"
	AGS  Mode = "AGS"
)

type Cpu struct {
	memory [8192]uint8 // 8-bit address bus, 8kb memory

	pc Word // program counter
	sp Word // stack pointer

	ir uint8 // interrupt register
	ie uint8 // interrupt enable

	// 8 bit bi-directional data bus
	databus chan uint8

	// 16 bit write only address bus
	addressbus chan uint16

	// accumulator & flag register
	af Word

	// general purpose register pairs
	bc Word
	de Word
	hl Word

	reg Registers
}

func (c *Cpu) fetch() uint8 {
	// read from memory
	data := c.memory[c.pc]
	c.pc++
	fmt.Println("Fetch")
	return data
}

func (c *Cpu) init(mode Mode) {
	switch mode {
	case CGB:
		c.af = 0x1100
		c.bc = 0x0100
		c.de = 0x0008
		c.hl = 0x007C
		c.sp = 0xFFFE
		c.pc = 0x0100
	}
}

func (c *Cpu) decode(opcode uint8) InstructionSet {

	// HALT
	if opcode == 0x76 {
	}

	// LD r8, r8
	if opcode >= 0x40 && opcode <= 0x7F {
		return InstructionSet{
			execute: ld_r8_r8,
			cycles:  1,
		}
	}

	return InstructionSet{
		execute: func(_ *Cpu, _ uint8) {},
		cycles:  1,
	}
}

func (c *Cpu) Start() {

	// init classic game-boy
	c.init(CGB)

	start := time.Now()
	ops := 0

	c.memory[0x0100] = 0x40

	// one cicle = 1us
	machineCycle := time.NewTicker(1 * time.Microsecond)

	stop := make(chan bool)

	go func() {
		var (
			remainingCycles int
			opcode          uint8
		)
		for {
			select {
			case <-stop:
				// force stop the cpu
				return
			case <-machineCycle.C:

				// if no opcode was read cycles is 0 (first cycle) or 1 (parallel fetch)
				if opcode == 0 && remainingCycles <= 1 {

					// fetch opcode from memory
					opcode = c.fetch()
					// we should spend 1 machine cycle during read
					remainingCycles = 1
				}

				// if opcode is 0, it means that we should not execute, otherwise, it means
				// that a cycle was spent fetching and now we should execute.
				// also we will only proceed after the number of required cycles
				// is passed (equals to 0)
				if opcode > 0 && remainingCycles == 0 {

					// decode - calculate how many cycles per instruction
					is := c.decode(uint8(opcode))

					// how many cycles for instruction
					remainingCycles = is.cycles

					fmt.Printf("required cycles %d\n", remainingCycles)

					// execute
					is.execute(c)

					// reset opcode
					opcode = 0x0
				}

				// clock observability
				ops++

				// decrease the number of cycles
				remainingCycles--
			}
		}
	}()

	time.Sleep(1 * time.Second)
	machineCycle.Stop()
	stop <- true

	fmt.Printf("Elapsed %d ops %d\n", time.Since(start).Milliseconds(), ops)
}
