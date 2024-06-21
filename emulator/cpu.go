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
	memory *Memory // 8-bit address bus, 8kb memory

	pc Word // program counter
	sp Word // stack pointer

	ir uint8 // interrupt register
	ie uint8 // interrupt enable

	requiredCycles int

	// general purpose register pairs
	reg Registers
}

func (c *Cpu) fetch() uint8 {
	// read from memory
	data := c.memory.Read(c.pc)
	c.pc++
	return data
}

func (c *Cpu) init(mode Mode) {
	switch mode {
	case CGB:
		c.reg.w8(reg_a, 0x11)
		c.reg.w_flag(0x0)
		c.reg.w16(reg_bc, 0x0100)
		c.reg.w16(reg_de, 0x0008)
		c.reg.w16(reg_hl, 0x007C)
		c.sp = 0xFFFE
		c.pc = 0x0100
	}
	c.memory = &Memory{}
}

func (c *Cpu) decode(opcode uint8) instruction {

	if opcode == 0x0 {
		// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#NOP
		return op_nop
	}

	// decode based on the gbdev instruction set 'BLOCKS' approach:
	// https://gbdev.io/pandocs/CPU_Instruction_Set.html

	// mask by 11000000 and right shift 6 to check how the two leftmost bits are set
	switch (opcode & 0xC0) >> 6 {

	/*
	 * BLOCK 0
	 */
	case 0x0:

		// mask by 1111 to check the rightmost nibble (four bits)
		switch opcode & 0xF {
		case 0x1:
			// ld r16, imm16
			return op_ld_r16_imm16
		case 0x2:
			// ld [r16mem], a
			return op_ld_r16mem_a
		case 0xA:
			// ld a, [r16mem]
			return op_ld_a_r16mem
		case 0x8:
			// ld [imm16], sp
			return op_ld_imm16_mem_sp
		case 0x3:
			// incr r16
			return op_inc_r16
		case 0xB:
			// dec r16
			return op_dec_r16
		case 0x9:
			// add hl, r16
			return op_add_hl_r16
		}

		// mask by 0111 to check how the rightmost 3 bits are set
		switch opcode & 0x7 {
		case 0x4:
			// inc r8
			return op_inc_r8
		case 0x5:
			// dec r8
			return op_dec_r8
		case 0x6:
			// ld r8, imm8
			return op_ld_r8_imm8
		case 0x7:

			// mask by 11111000 and right shift by 3 to ignore the rightmost 3 bits
			switch (opcode & 0xF8) >> 3 {

			case 0x0:
				// rlca
				return op_rotate_rlca
			case 0x1:
				// rrca
				return op_rotate_rrca
			case 0x2:
				// rla
				return op_rotate_rla
			case 0x3:
				// rra
				return op_rotate_rra
			case 0x4:
				// daa
				return op_daa
			case 0x5:
				// cpl
				return op_cpl
			case 0x6:
				// scf
				return op_scf
			case 0x7:
				// ccf
				return op_ccf
			}

		case 0x0:
			// since the 3 rightmost bits are 0, we can safely RSH by 3 to get the real value
			switch opcode >> 3 {
			case 0x2:
				// stop
				// https://gist.github.com/SonoSooS/c0055300670d678b5ae8433e20bea595#nop-and-stop
				return op_nop
			case 0x3:
				// jr imm8
				return op_jr_imm8
			default:
				// jr cond, imm8
				return op_jr_cond_imm8
			}
		}

	/*
	 * BLOCK 1
	 */
	case 0x1:
		// mask by 00111111 to get the rightmost 6 bits
		switch opcode & 0x3F {
		case 0x36:
			// halt
			// https://gbdev.io/pandocs/halt.html#halt
		default:
			// ld r8, r8
			return op_ld_r8_r8
		}

	/*
	 * BLOCK 2
	 */
	case 0x2:

		// get the bits from 5 to 3 (left to right)
		switch (opcode & 0x38) >> 3 {
		case 0x0:
			// add a, r8
			return op_add_a_r8
		case 0x1:
			// adc a, r8
			return op_adc_a_r8
		case 0x2:
			// sub a, r8
			return op_sub_a_r8
		case 0x3:
			// sbc a, r8
			return op_sbc_a_r8
		case 0x4:
			// and a, r8
			return op_and_a_r8
		case 0x5:
			// xor a, r8
			return op_xor_a_r8
		case 0x6:
			// or a, r8
			return op_or_a_r8
		case 0x7:
			// cp a, r8
			return op_cp_a_r8
		}

	/*
	 * BLOCK 3
	 */
	case 0x3:

		// check the three rightmost bits
		switch opcode & 0x7 {

		case 0x6:
			// get the bits from 5 to 3 (left to right)
			switch (opcode & 0x38) >> 3 {
			case 0x0:
				// add a, imm8
				return op_add_a_imm8
			case 0x1:
				// adc a, imm8
				return op_adc_a_imm8
			case 0x2:
				// sub a, imm8
				return op_sub_a_imm8
			case 0x3:
				// sbc a, imm8
				return op_sbc_a_imm8
			case 0x4:
				// and a, imm8
				return op_and_a_imm8
			case 0x5:
				// xor a, imm8
				return op_xor_a_imm8
			case 0x6:
				// or a, imm8
				return op_or_a_imm8
			case 0x7:
				// cp a, imm8
				return op_cp_a_imm8
			}
		}
	}

	// default is nop, might check for errors later
	return op_nop
}

func (c *Cpu) Start() {

	// init classic game-boy
	c.init(CGB)

	start := time.Now()
	ops := 0

	c.memory.Write(0x0100, 0x40)

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
					remainingCycles = c.requiredCycles

					fmt.Printf("required cycles %d\n", remainingCycles)

					// execute
					is(c, opcode)

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
