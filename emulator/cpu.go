package emulator

import (
	"fmt"
	"math/rand"
	"time"
)

type Cpu struct {
	memory [4096]Word // 16-bit address bus, 8kb memory

	pc Word // program counter
	sp Word // stack pointer

	a  uint8 // accumulator
	f  uint8 // flag register
	ir uint8 // interrupt register
	ie uint8 // interrupt enable

	// 8 bit bi-directional data bus
	databus chan uint8

	// 16 bit write only address bus
	addressbus chan uint16

	// general purpose register pairs
	af Word
	bc Word
	de Word
	hl Word
}

func (c *Cpu) fetch() Word {
	// read from memory
	data := c.memory[c.pc]
	c.pc++
	fmt.Println("Fetch")
	return Word(data)
}

func (c *Cpu) execute(opcode Word) {
	fmt.Println("Execute")
}

func (c *Cpu) Start() {

	start := time.Now()
	ops := 0

	for i := range c.memory {
		c.memory[i] = 0x1
	}

	c.pc = Word(0x0000)

	// one cicle = 1us
	machineCycle := time.NewTicker(1 * time.Microsecond)

	stop := make(chan bool)

	go func() {
		var (
			remainingCycles int
			opcode          Word
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
					remainingCycles = rand.Intn(4) + 1
					fmt.Printf("required cycles %d\n", remainingCycles)

					// execute in parallel
					c.execute(opcode)

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
