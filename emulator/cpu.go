package emulator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed opcodes.json
var opcodesFile []byte

type Mode string

const (
	INTERRUPT_ENABLE = Word(0xFFFF)
	INTERRUPT_FLAG   = Word(0xFF0F)
	// https://gbdev.io/pandocs/The_Cartridge_Header.html#0100-0103--entry-point
	CPU_START = Word(0x0100)
)

const (
	STEP_QUIT = iota
	STEP_RESUME
	STEP_NEXT
)

const (
	DMG  Mode = "DMG"
	MGB  Mode = "MGB"
	SGB  Mode = "SGB"
	SGB2 Mode = "SGB2"
	CGB  Mode = "CGB"
	AGB  Mode = "AGB"
	AGS  Mode = "AGS"
)

// Opcodes used for debugging purposes
type Opcodes struct {
	Unprefixed map[string]json.RawMessage `json:"unprefixed"`
	Cbprefixed map[string]json.RawMessage `json:"cbprefixed"`
}

type Cpu struct {
	memory *Memory // 8-bit address bus

	previousPC Word // debugging
	pc         Word // program counter
	sp         Word // stack pointer

	ime      uint8 // interrupt master enable
	enableEI bool  // delay EI

	remainingCycles int
	opcode          uint8

	requiredCycles int

	scheduledSerial int
	scheduledOAMDma int
	oamDmaSource    int

	// general purpose register pairs
	reg Registers

	haltBug bool
	halted  bool

	debug       bool
	step        bool
	profiling   bool
	stopped     bool
	breakPoints string
	cbprefixed  bool
	silent      bool
	opcodes     *Opcodes
}

func (c *Cpu) fetch() uint8 {
	// read from memory
	data := c.memory.Read(c.pc)
	c.previousPC = c.pc
	c.pc++
	return data
}

func (c *Cpu) setup(mode Mode) {
	switch mode {
	case DMG:
		c.reg.w_flag(0x0)
		c.reg.w8(reg_a, 0x01)
		c.reg.w8(reg_f, 0xB0)
		c.reg.w16(reg_bc, 0x0013)
		c.reg.w16(reg_de, 0x00D8)
		c.reg.w16(reg_hl, 0x014D)
		c.sp = 0xFFFE
		c.pc = CPU_START

		// https://gbdev.io/pandocs/Power_Up_Sequence.html#hardware-registers
		c.memory.Write(LCDC_REGISTER, 0x91)
		c.memory.Write(LCD_REGISTER, 0x85)
		c.memory.Write(INTERRUPT_FLAG, 0xE1)
		c.memory.Write(PORT_SERIAL_TRANSFER_SC, 0x7E)
		c.memory.Write(PORT_SERIAL_TRANSFER_SB, 0x00)
		c.memory.mem[PORT_JOYPAD] = 0xCF // must bypass write method
		c.memory.mem[NR10] = 0x80
		c.memory.mem[NR11] = 0xBF
		c.memory.mem[NR12] = 0xF3
		c.memory.mem[NR13] = 0xFF
		c.memory.mem[NR14] = 0xBF
		c.memory.mem[NR21] = 0x3F
		c.memory.mem[NR22] = 0x00
		c.memory.mem[NR23] = 0xFF
		c.memory.mem[NR24] = 0xBf
		c.memory.mem[NR30] = 0x7F
		c.memory.mem[NR31] = 0xFF
		c.memory.mem[NR32] = 0x9F
		c.memory.mem[NR33] = 0xFF
		c.memory.mem[NR34] = 0xBF
		c.memory.mem[NR41] = 0xFF
		c.memory.mem[NR42] = 0x00
		c.memory.mem[NR43] = 0x00
		c.memory.mem[NR44] = 0xBF
		c.memory.mem[NR50] = 0x77
		c.memory.mem[NR51] = 0xF3
		c.memory.mem[NR52] = 0xF1
	}
}

func (c *Cpu) decode(opcode uint8) (instruction, string) {

	// reset
	c.cbprefixed = false

	if opcode == 0x0 {
		// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#NOP
		return op_nop, "nop"
	}

	// decode based on the gbdev instruction set 'BLOCKS' approach:
	// https://gbdev.io/pandocs/CPU_Instruction_Set.html

	// mask by 11000000 and right shift 6 to check how the two leftmost bits are set
	switch (opcode & 0xC0) >> 6 {

	/*
	 * BLOCK 0
	 */
	case 0x0:
		if c.debug {
			log.Println("BLOCK 0")
		}

		// mask by 1111 to check the rightmost nibble (four bits)
		switch opcode & 0xF {
		case 0x1:
			// ld r16, imm16
			return op_ld_r16_imm16, "ld r16, imm16"
		case 0x2:
			// ld [r16mem], a
			return op_ld_r16mem_a, "ld [r16mem], a"
		case 0xA:
			// ld a, [r16mem]
			return op_ld_a_r16mem, "ld a, [r16mem]"
		case 0x8:
			if opcode&0xF0 == 0 {
				// ld [imm16], sp
				return op_ld_imm16_mem_sp, "ld [imm16], sp"
			}
		case 0x3:
			// incr r16
			return op_inc_r16, "incr r16"
		case 0xB:
			// dec r16
			return op_dec_r16, "dec r16"
		case 0x9:
			// add hl, r16
			return op_add_hl_r16, "add hl, r16"
		}

		// mask by 0111 to check how the rightmost 3 bits are set
		switch opcode & 0x7 {
		case 0x4:
			// inc r8
			return op_inc_r8, "inc r8"
		case 0x5:
			// dec r8
			return op_dec_r8, "dec r8"
		case 0x6:
			// ld r8, imm8
			return op_ld_r8_imm8, "ld r8, imm8"
		case 0x7:

			// mask by 11111000 and right shift by 3 to ignore the rightmost 3 bits
			switch (opcode & 0xF8) >> 3 {

			case 0x0:
				// rlca
				return op_rotate_rlca, "rlca"
			case 0x1:
				// rrca
				return op_rotate_rrca, "rrca"
			case 0x2:
				// rla
				return op_rotate_rla, "rla"
			case 0x3:
				// rra
				return op_rotate_rra, "rra"
			case 0x4:
				// daa
				return op_daa, "daa"
			case 0x5:
				// cpl
				return op_cpl, "cpl"
			case 0x6:
				// scf
				return op_scf, "scf"
			case 0x7:
				// ccf
				return op_ccf, "ccf"
			}

		case 0x0:
			// since the 3 rightmost bits are 0, we can safely RSH by 3 to get the real value
			switch opcode >> 3 {
			case 0x2:
				// reset timer
				c.memory.Write(PORT_DIV, 0x0)
				// stop
				// https://gist.github.com/SonoSooS/c0055300670d678b5ae8433e20bea595#nop-and-stop
				return op_nop, "stop"
			case 0x3:
				// jr imm8
				return op_jr_imm8, "jr imm8"
			default:
				// jr cond, imm8
				return op_jr_cond_imm8, "jr cond, imm8"
			}
		}

	/*
	 * BLOCK 1
	 */
	case 0x1:
		if c.debug {
			log.Println("BLOCK 1")
		}
		// mask by 00111111 to get the rightmost 6 bits
		switch opcode & 0x3F {
		case 0x36:
			// halt
			// https://gbdev.io/pandocs/halt.html#halt
			return op_halt, "op_halt"

		default:
			// ld r8, r8
			return op_ld_r8_r8, "ld r8, r8"
		}

	/*
	 * BLOCK 2
	 */
	case 0x2:
		if c.debug {
			log.Println("BLOCK 2")
		}
		// get the bits from 5 to 3 (left to right)
		switch (opcode & 0x38) >> 3 {
		case 0x0:
			// add a, r8
			return op_add_a_r8, "add a, r8"
		case 0x1:
			// adc a, r8
			return op_adc_a_r8, "adc a, r8"
		case 0x2:
			// sub a, r8
			return op_sub_a_r8, "sub a, r8"
		case 0x3:
			// sbc a, r8
			return op_sbc_a_r8, "sbc a, r8"
		case 0x4:
			// and a, r8
			return op_and_a_r8, "and a, r8"
		case 0x5:
			// xor a, r8
			return op_xor_a_r8, "xor a, r8"
		case 0x6:
			// or a, r8
			return op_or_a_r8, "or a, r8"
		case 0x7:
			// cp a, r8
			return op_cp_a_r8, "cp a, r8"
		}

	/*
	 * BLOCK 3
	 */
	case 0x3:
		if c.debug {
			log.Println("BLOCK 3")
		}
		// get the bits from 5 to 3 (left to right)
		switch opcode {
		case 0xC3:
			// jp imm16
			return op_jp_imm16, "jp imm16"
		case 0xC6:
			// add a, imm8
			return op_add_a_imm8, "add a, imm8"
		case 0xC9:
			// ret
			return op_ret, "ret"
		case 0xCB:
			// CB prefixed
			c.cbprefixed = true

			// https://gbdev.io/pandocs/CPU_Instruction_Set.html#cb-prefix-instructions
			prefix := c.fetch()
			c.opcode = prefix

			switch (prefix & 0xF8) >> 3 {
			case 0x0:
				// rlc r8
				return op_rlc_r8, "CB rlc r8"
			case 0x1:
				// rrc r8
				return op_rrc_r8, "CB rrc r8"
			case 0x2:
				// rl r8
				return op_rl_r8, "CB rl r8"
			case 0x3:
				// rr r8
				return op_rr_r8, "CB rr r8"
			case 0x4:
				// sla r8
				return op_sla_r8, "CB sla r8"
			case 0x5:
				// sra r8
				return op_sra_r8, "CB sra r8"
			case 0x6:
				// swap r8
				return op_swap_r8, "CB swap r8"
			case 0x7:
				// srl r8
				return op_srl_r8, "CB srl r8"
			}

			switch (prefix & 0xC0) >> 6 {
			case 0x1:
				// bit b3, r8
				return op_bit_r8, "CB bit b3, r8"
			case 0x2:
				// res b3, r8
				return op_res_r8, "CB res b3, r8"
			case 0x3:
				// set b3, r8
				return op_set_r8, "CB set b3, r8"
			}

		case 0xCD:
			// call imm16
			return op_call_imm16, "call imm16"
		case 0xCE:
			// adc a, imm8
			return op_adc_a_imm8, "adc a, imm8"
		case 0xD6:
			// sub a, imm8
			return op_sub_a_imm8, "sub a, imm8"
		case 0xD9:
			// reti
			return op_reti, "reti"
		case 0xDE:
			// sbc a, imm8
			return op_sbc_a_imm8, "sbc a, imm8"
		case 0xE0:
			// ldh [imm8], a
			return op_ldh_imm8_a, "ldh [imm8], a"
		case 0xE2:
			// ldh [c], a
			return op_ldh_c_a, "ldh [c], a"
		case 0xE6:
			// and a, imm8
			return op_and_a_imm8, "and a, imm8"
		case 0xE8:
			// add sp, imm8
			return op_add_sp_imm8, "add sp, imm8"
		case 0xE9:
			// jp hl
			return op_jp_hl, "jp hl"
		case 0xEA:
			// ld [n16], a
			return op_ld_imm16_a, "ld [n16], a"
		case 0xEE:
			// xor a, imm8
			return op_xor_a_imm8, "xor a, imm8"
		case 0xF0:
			// ldh a, [imm8]
			return op_ldh_a_imm8, "ldh a, [imm8]"
		case 0xF2:
			// ld a, [c]
			return op_ldh_a_c, "ld a, [c]"
		case 0xF3:
			// di
			return op_di, "di"
		case 0xF6:
			// or a, imm8
			return op_or_a_imm8, "or a, imm8"
		case 0xF8:
			// ld hl, SP+e8
			return op_ld_sp_e, "ld hl, SP+e8"
		case 0xF9:
			// ld sp, hl
			return op_ld_sp_hl, "ld sp, hl"
		case 0xFA:
			// ld a, [imm16]
			return op_ld_a_imm16, "ld a, [imm16]"
		case 0xFB:
			// ei
			return op_ei, "ei"
		case 0xFE:
			// cp a, imm8
			return op_cp_a_imm8, "cp a, imm8"
		}

		// since we are inside block 3, mask to get the 3 rightmost bits
		switch opcode & 0x7 {
		case 0x0:
			// ret cond
			return op_ret_cond, "ret cond"
		case 0x2:
			// jp cond, imm16
			return op_jp_cond, "jp cond, imm16"
		case 0x4:
			// call cond, imm16
			return op_call_cond, "call cond, imm16"
		case 0x7:
			// rst tgt3
			return op_rst, "rst tgt3"
		}

		// also try to mask the 4 rightmost bits
		switch opcode & 0xf {

		case 0x1:
			// pop r16stk
			return op_pop_r16stk, "pop r16stk"
		case 0x5:
			// push r16stk
			return op_push_r16stk, "push r16stk"
		}
	}

	// default is nop, might check for errors later
	return op_nop, "unknown nop"
}

func (c *Cpu) pushPCIntoStack() {
	pc := c.pc
	c.sp--
	c.memory.Write(c.sp, pc.High())
	c.sp--
	c.memory.Write(c.sp, pc.Low())
}

func (c *Cpu) interruptRequested() bool {

	requested := false
	ienable := c.memory.Read(INTERRUPT_ENABLE)
	iflag := c.memory.Read(INTERRUPT_FLAG)

	if c.ime == 0 {
		return false
	}

	if (iflag&ienable)&0x1 > 0 {

		if c.debug {
			log.Printf("VBLANK INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)
		}

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// vBlank
		c.previousPC = c.pc
		c.pc = Word(0x40)

		// unset
		iflag &= 0xFE
		requested = true

	} else if (iflag&ienable)&0x2 > 0 {

		if c.debug {
			log.Printf("LCDC/STAT INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)
		}

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// LCDC / STAT
		c.previousPC = c.pc
		c.pc = Word(0x48)

		// unset
		iflag &= 0xFD
		requested = true

	} else if (iflag&ienable)&0x4 > 0 {

		if c.debug {
			log.Printf("TIMER INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)
		}

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// TIMER
		c.previousPC = c.pc
		c.pc = Word(0x50)

		// unset
		iflag &= 0xFB
		requested = true

	} else if (iflag&ienable)&0x8 > 0 {

		if c.debug {
			log.Printf("SERIAL INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)
		}

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// SERIAL
		c.previousPC = c.pc
		c.pc = Word(0x58)

		// unset
		iflag &= 0xF7
		requested = true

	} else if (iflag&ienable)&0x10 > 0 {

		log.Printf("JOYPAD INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// JOYPAD
		c.previousPC = c.pc
		c.pc = Word(0x60)

		// unset
		iflag &= 0xEF
		requested = true
	}

	if !requested {
		return false
	}

	if c.debug {
		log.Printf("INTERRUPT ACKNOWLEDGED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)
	}

	// acknowledge
	c.memory.Write(INTERRUPT_FLAG, iflag)

	// disable interrupts
	c.ime = 0

	return requested
}

func (c *Cpu) sync(cycle int) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("FAULT AT OP (val=0x%X bit=%.8b) CYCLE=%d REMAINING=%d PC=0x%X SP=0x%X IME=%d HL=0x%X\n", c.opcode, c.opcode, cycle, c.remainingCycles, c.pc, c.sp, c.ime, c.reg.r16(reg_hl))
			log.Fatalf("%v : %s\n", r, string(debug.Stack()))
		}
	}()

	sb := c.memory.Read(PORT_SERIAL_TRANSFER_SB)
	sc := c.memory.Read(PORT_SERIAL_TRANSFER_SC)

	if !c.halted && sc&0x80 > 0 && sc&0x1 > 0 { // transfer enabled AND clock master

		modulo := 80 // 262144 Hz
		if sc&0x2 > 0 {
			modulo = 16
		}

		if cycle%modulo == 0 {
			if c.scheduledSerial == 0 {
				log.Printf("RECEIVED %d (0x%.8X) ROM SERIAL (PC=0x%.8X)\n", sb, sb, c.pc)
				c.scheduledSerial = 8
			}

			if c.scheduledSerial > 0 {
				c.memory.Write(PORT_SERIAL_TRANSFER_SB, sb<<1) // incoming data
				c.scheduledSerial--
			}

			if c.scheduledSerial == 0 {
				c.memory.Write(PORT_SERIAL_TRANSFER_SB, 0x0)     // clear SB
				c.memory.Write(PORT_SERIAL_TRANSFER_SC, sc&0x7F) // clear bit 7
				// request SERIAL interruption
				c.memory.Write(INTERRUPT_FLAG, c.memory.Read(INTERRUPT_FLAG)|0x8)
			}
		}
	}

	if !c.halted && c.scheduledOAMDma >= 0 && !c.memory.dma {
		if c.scheduledOAMDma < 160 {
			rSrcAddr := NewWord(uint8(c.oamDmaSource), uint8(159-c.scheduledOAMDma))
			rTgtAddr := NewWord(0xFE, uint8(159-c.scheduledOAMDma))
			rVal := c.memory.Read(rSrcAddr)
			if c.debug {
				log.Printf("COPY OAM DMA 0x%X FROM 0x%X TO 0x%X (PC=0x%.8X, IDX=%d)\n", rVal, rSrcAddr, rTgtAddr, c.pc, c.scheduledOAMDma)
			}
			c.memory.Write(rTgtAddr, rVal)
		}
		c.scheduledOAMDma--
	}

	// halt AND interrupt is pending
	interruptPending := (c.memory.Read(INTERRUPT_ENABLE) & c.memory.Read(INTERRUPT_FLAG)) > 0
	if c.halted && interruptPending {
		if c.debug {
			log.Printf("HALT RESUMED bug=%t ime=%d pending=%t\n", c.haltBug, c.ime, interruptPending)
		}
		c.halted = false
	}

	// if no opcode was read cycles is 0 (first cycle) or 1 (parallel fetch)
	if !c.halted && c.opcode == 0 && c.remainingCycles <= 1 {

		// Interrupts are accepted during the op code fetch cycle of each instruction
		if c.interruptRequested() {
			ienable := c.memory.Read(INTERRUPT_ENABLE)
			iflag := c.memory.Read(INTERRUPT_FLAG)
			if c.debug {
				log.Printf("INTERRUPT PC=0x%X SP=0x%X IME=%d IE=%.8b IF=%.8b\n", c.pc, c.sp, c.ime, ienable, iflag)
			}
		}

		// fetch opcode from memory
		c.opcode = c.fetch()

		// https://gbdev.io/pandocs/halt.html#halt-bug
		if c.haltBug {
			c.pc--
			c.haltBug = false
		}

		c.remainingCycles = 1
	}

	// if opcode is 0, it means that we should not execute, otherwise, it means
	// that a cycle was spent fetching and now we should execute.
	// also we will only proceed after the number of required cycles
	// is passed (equals to 0)
	if !c.halted && c.opcode > 0 && c.remainingCycles == 0 {

		if c.enableEI {
			c.ime = 1
			if c.debug {
				log.Println("ENABLE EI")
			}
			c.enableEI = false
		}

		// decode - calculate how many cycles per instruction
		is, operation := c.decode(uint8(c.opcode))

		if c.debug || operation == "unknown nop" {
			if c.cbprefixed {
				op, ok := c.opcodes.Cbprefixed[fmt.Sprintf("0x%.2X", c.opcode)]
				if ok {
					log.Printf("CB %s", string(op))
				}
			} else {
				op, ok := c.opcodes.Unprefixed[fmt.Sprintf("0x%.2X", c.opcode)]
				if ok {
					log.Printf("%s", string(op))
				}
			}
			if operation == "unknown nop" {
				c.step = true
			}
		}

		if c.shouldStep(c.opcode, operation) {
			var sb strings.Builder
			sb.WriteString("STEP BEFORE\n")
			sb.WriteString("*********************************\n")

			sb.WriteString(fmt.Sprintf("\tOP (val=0x%X bit=%.8b name=%s) \n\tPC=0x%X\n\tSP=0x%X\n\tIME=%d\n\tPREV_PC=0x%X\n", c.opcode, c.opcode, operation, c.pc-1, c.sp, c.ime, c.previousPC))

			sb.WriteString(fmt.Sprintf("\tREG_B=0x%X\n\tREG_C=0x%X\n\tREG_D=0x%X\n\tREG_E=0x%X\n\tREG_H=0x%X\n\tREG_L=0x%X\n\tREG_A=0x%X\n\tREG_F=0x%X\n", c.reg.r8(reg_b), c.reg.r8(reg_c), c.reg.r8(reg_d), c.reg.r8(reg_e), c.reg.r8(reg_h), c.reg.r8(reg_l), c.reg.r8(reg_a), c.reg.r8(reg_f)))

			sb.WriteString(fmt.Sprintf("\tREG_BC=0x%X\n\tREG_DE=0x%X\n\tREG_HL=0x%X\n\tREG_SP=0x%X\n", c.reg.r16(reg_bc), c.reg.r16(reg_de), c.reg.r16(reg_hl), c.sp))

			sb.WriteString(fmt.Sprintf("\tZ_FLAG=%d\n\tN_FLAG=%d\n\tC_FLAG=%d\n\tH_FLAG=%d\n", c.reg.r_flags()&z_flag, c.reg.r_flags()&n_flag, c.reg.r_flags()&c_flag, c.reg.r_flags()&h_flag))

			sb.WriteString("*********************************\n")

			log.Print(sb.String())

			switch c.waitStep() {
			case STEP_QUIT:
				c.stopped = true
				return
			case STEP_RESUME:
				c.step = false
			}
		}

		// execute
		is(c, c.opcode)

		// how many cycles for instruction
		c.remainingCycles = c.requiredCycles

		if c.shouldStep(c.opcode, operation) {
			var sb strings.Builder
			sb.WriteString("STEP AFTER\n")
			sb.WriteString("*********************************\n")

			sb.WriteString(fmt.Sprintf("\tOP (val=0x%X bit=%.8b name=%s) \n\tPC=0x%X\n\tSP=0x%X\n\tIME=%d\n\tPREV_PC=0x%X\n", c.opcode, c.opcode, operation, c.pc, c.sp, c.ime, c.previousPC))

			sb.WriteString(fmt.Sprintf("\tREG_B=0x%X\n\tREG_C=0x%X\n\tREG_D=0x%X\n\tREG_E=0x%X\n\tREG_H=0x%X\n\tREG_L=0x%X\n\tREG_A=0x%X\n\tREG_F=0x%X\n", c.reg.r8(reg_b), c.reg.r8(reg_c), c.reg.r8(reg_d), c.reg.r8(reg_e), c.reg.r8(reg_h), c.reg.r8(reg_l), c.reg.r8(reg_a), c.reg.r8(reg_f)))

			sb.WriteString(fmt.Sprintf("\tREG_BC=0x%X\n\tREG_DE=0x%X\n\tREG_HL=0x%X\n\tREG_SP=0x%X\n", c.reg.r16(reg_bc), c.reg.r16(reg_de), c.reg.r16(reg_hl), c.sp))

			sb.WriteString(fmt.Sprintf("\tZ_FLAG=%d\n\tN_FLAG=%d\n\tC_FLAG=%d\n\tH_FLAG=%d\n", c.reg.r_flags()&z_flag, c.reg.r_flags()&n_flag, c.reg.r_flags()&c_flag, c.reg.r_flags()&h_flag))

			sb.WriteString("*********************************\n")

			log.Print(sb.String())

			switch c.waitStep() {
			case STEP_QUIT:
				return
			case STEP_RESUME:
				c.step = false
			}
		} else if !c.silent {
			log.Printf("OP (val=0x%X bit=%.8b name=%s) CYCLE=%d REMAINING=%d PC=0x%X SP=0x%X IME=%d HL=0x%X\n", c.opcode, c.opcode, operation, cycle, c.remainingCycles, c.pc-1, c.sp, c.ime, c.reg.r16(reg_hl))
		}

		if operation == "stop" {
			log.Println("STOP INSTRUCTION RECEIVED")
			c.stopped = true
			return
		}

		if c.remainingCycles == 1 {

			// Interrupts are accepted during the op code fetch cycle of each instruction
			if c.interruptRequested() && c.debug {
				log.Printf("INTERRUPT PC=0x%X SP=0x%X IME=%d\n", c.pc, c.sp, c.ime)
			}

			// fetch opcode from memory
			c.opcode = c.fetch()

		} else {
			// reset opcode
			c.opcode = 0x0
		}
	}

	// handle OAM request
	oam := int(c.memory.Read(PORT_OAM_DMA_CONTROL))
	if !c.halted && c.memory.dma && c.scheduledOAMDma < 0 {
		if c.debug {
			log.Printf("SCHEDULED OAM DMA AT 0x%X (PC=0x%.8X)\n", oam, c.pc)
		}
		c.scheduledOAMDma = 160 // 1 m-cycle delay
		c.oamDmaSource = oam
		c.memory.dma = false // acknowledge DMA
	}

	if !c.halted {

		// decrease the number of cycles
		c.remainingCycles--
	}
}

func (c *Cpu) init() error {

	// init classic game-boy
	c.setup(DMG)

	var opcodes Opcodes
	if err := json.Unmarshal(opcodesFile, &opcodes); err != nil {
		return err
	}
	c.opcodes = &opcodes
	return nil
}

func (c *Cpu) shouldStep(opcode uint8, operation string) bool {

	if c.step {
		return true
	}

	bg := strings.TrimSpace(c.breakPoints)
	if !c.step && bg == "" {
		return false
	}

	for _, instruction := range strings.Split(bg, ";") {

		parts := strings.SplitN(instruction, "=", 2)
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "PC":
			n, err := strconv.ParseInt(parts[1], 16, 32)
			if err != nil {
				log.Printf("Invalid instruction filter syntax : %s\n", err.Error())
			}
			match := c.pc-1 == Word(n)
			if match {
				c.breakPoints = strings.Replace(c.breakPoints, instruction, "", 1)
				c.step = true
				c.memory.Write(0xFFFF, 1)
			}
			return match
		case "OPC":
			n, err := strconv.ParseInt(parts[1], 16, 32)
			if err != nil {
				log.Printf("Invalid instruction filter syntax : %s\n", err.Error())
			}
			match := opcode == uint8(n)
			if match {
				c.breakPoints = strings.Replace(c.breakPoints, instruction, "", 1)
				c.step = true
			}
			return match
		case "OPN":
			match := strings.EqualFold(operation, parts[1])
			if match {
				c.breakPoints = strings.Replace(c.breakPoints, instruction, "", 1)
				c.step = true
			}
			return match
		}
	}

	return false
}

func (c *Cpu) waitStep() int {
	for {

		key := rl.GetKeyPressed()
		if key == 0 {
			continue
		}

		switch key {
		case rl.KeyD:
			c.debug = !c.debug
		case rl.KeyEscape:
		case rl.KeyQ:
			return STEP_QUIT
		case rl.KeyN:
			return STEP_NEXT
		case rl.KeyR:
			return STEP_RESUME
		}
	}
}
