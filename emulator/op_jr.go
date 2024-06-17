package emulator

import "fmt"

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JR_n16
func op_jr_imm8(c *Cpu, _ uint8) {

	// m-cycles = 3
	c.requiredCycles = 3

	// retrieve N from memory and increase PC
	offset := c.fetch()

	fmt.Printf("PC=%X\n", c.pc)

	// check sign bit
	signed := ((offset & 0x80) >> 7) > 0

	// a
	// b
	// c -2
	if signed {
		offset = ^offset + 1
		fmt.Printf("%X\n", offset)
		c.pc -= Word(offset)
	} else {
		c.pc += Word(offset)
	}

	fmt.Printf("PC=%X\n", c.pc)
}
