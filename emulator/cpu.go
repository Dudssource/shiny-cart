package emulator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
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

	pc Word // program counter
	sp Word // stack pointer

	ime uint8 // interrupt master enable

	remainingCycles int
	opcode          uint8

	requiredCycles int

	// general purpose register pairs
	reg Registers

	doneChan chan bool
	stopChan chan bool

	debug       bool
	step        bool
	stopped     bool
	breakPoints string
	cbprefixed  bool
	silent      bool
	opcodes     *Opcodes
}

func (c *Cpu) fetch() uint8 {
	// read from memory
	data := c.memory.Read(c.pc)
	c.pc++
	return data
}

func (c *Cpu) setup(mode Mode) {
	switch mode {
	case CGB:
		c.reg.w8(reg_a, 0x11)
		c.reg.w_flag(0x0)
		c.reg.w16(reg_bc, 0x0100)
		c.reg.w16(reg_de, 0x0008)
		c.reg.w16(reg_hl, 0x007C)
		c.sp = 0xFFFE
		c.pc = CPU_START
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

			switch prefix & 0xF8 {
			case 0x0:
				// rlc r8
				return op_rlc_r8, "CB rlc r8"
			case 0x1:
				// rrc r8
				return op_rrc_r8, "CB rrc r8"
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

		log.Printf("VBLANK INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// vBlank
		c.pc = Word(0x40)

		// unset
		iflag &= 0xFE
		requested = true

	} else if (iflag&ienable)&0x2 > 0 {

		log.Printf("LCDC/STAT INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// LCDC / STAT
		c.pc = Word(0x48)

		// unset
		iflag &= 0xFD
		requested = true

	} else if (iflag&ienable)&0x4 > 0 {

		log.Printf("TIMER INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// TIMER
		c.pc = Word(0x50)

		// unset
		iflag &= 0xFB
		requested = true

	} else if (iflag&ienable)&0x8 > 0 {

		log.Printf("SERIAL INTERRUPT REQUESTED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)

		// 4 cycles to begin the interruption
		c.requiredCycles = 4

		// push PC into stack
		c.pushPCIntoStack()

		// SERIAL
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
		c.pc = Word(0x60)

		// unset
		iflag &= 0xEF
		requested = true
	}

	if !requested {
		return false
	}

	log.Printf("INTERRUPT ACKNOWLEDGED IFLAG=%.8b IENABLE=%.8b\n", iflag, ienable)

	// acknowledge
	c.memory.Write(INTERRUPT_FLAG, iflag)

	// disable interrupts
	c.ime = 0

	return requested
}

func (c *Cpu) sync(cycle int) {
	// if no opcode was read cycles is 0 (first cycle) or 1 (parallel fetch)
	if c.opcode == 0 && c.remainingCycles <= 1 {

		// Interrupts are accepted during the op code fetch cycle of each instruction
		if c.interruptRequested() {

			// how many cycles for instruction
			c.remainingCycles = c.requiredCycles

			log.Printf("INTERRUPT PC=%X SP=%X IME=%d\n", c.pc, c.sp, c.ime)

			// nop
			return
		}

		// fetch opcode from memory
		c.opcode = c.fetch()

		// we should spend 1 machine cycle during read
		c.remainingCycles = 1
	}

	// if opcode is 0, it means that we should not execute, otherwise, it means
	// that a cycle was spent fetching and now we should execute.
	// also we will only proceed after the number of required cycles
	// is passed (equals to 0)
	if c.opcode > 0 && c.remainingCycles == 0 {

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

		if c.shouldStep() {
			var sb strings.Builder
			sb.WriteString("STEP BEFORE\n")
			sb.WriteString("*********************************\n")

			sb.WriteString(fmt.Sprintf("\tOP (val=%X bit=%.8b name=%s) \n\tPC=%X\n\tSP=%X\n\tIME=%d\n", c.opcode, c.opcode, operation, c.pc, c.sp, c.ime))

			sb.WriteString(fmt.Sprintf("\tREG_B=%X\n\tREG_C=%X\n\tREG_D=%X\n\tREG_E=%X\n\tREG_H=%X\n\tREG_L=%X\n\tREG_A=%X\n\tREG_F=%X\n", c.reg.r8(reg_b), c.reg.r8(reg_c), c.reg.r8(reg_d), c.reg.r8(reg_e), c.reg.r8(reg_h), c.reg.r8(reg_l), c.reg.r8(reg_a), c.reg.r8(reg_f)))

			sb.WriteString(fmt.Sprintf("\tREG_BC=%X\n\tREG_DE=%X\n\tREG_HL=%X\n\tREG_SP=%X\n", c.reg.r16(reg_bc), c.reg.r16(reg_de), c.reg.r16(reg_hl), c.reg.r16(reg_sp)))

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

		if c.shouldStep() {
			var sb strings.Builder
			sb.WriteString("STEP AFTER\n")
			sb.WriteString("*********************************\n")

			sb.WriteString(fmt.Sprintf("\tOP (val=%X bit=%.8b name=%s) \n\tPC=%X\n\tSP=%X\n\tIME=%d\n", c.opcode, c.opcode, operation, c.pc, c.sp, c.ime))

			sb.WriteString(fmt.Sprintf("\tREG_B=%X\n\tREG_C=%X\n\tREG_D=%X\n\tREG_E=%X\n\tREG_H=%X\n\tREG_L=%X\n\tREG_A=%X\n\tREG_F=%X\n", c.reg.r8(reg_b), c.reg.r8(reg_c), c.reg.r8(reg_d), c.reg.r8(reg_e), c.reg.r8(reg_h), c.reg.r8(reg_l), c.reg.r8(reg_a), c.reg.r8(reg_f)))

			sb.WriteString(fmt.Sprintf("\tREG_BC=%X\n\tREG_DE=%X\n\tREG_HL=%X\n\tREG_SP=%X\n", c.reg.r16(reg_bc), c.reg.r16(reg_de), c.reg.r16(reg_hl), c.reg.r16(reg_sp)))

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
			log.Printf("OP (val=%X bit=%.8b name=%s) CYCLE=%d REMAINING=%d PC=%X SP=%X IME=%d HL=%X\n", c.opcode, c.opcode, operation, cycle, c.remainingCycles, c.pc, c.sp, c.ime, c.reg.r16(reg_hl))
		}

		if operation == "stop" {
			log.Println("STOP INSTRUCTION RECEIVED")
			return
		}

		// reset opcode
		c.opcode = 0x0
	}

	// decrease the number of cycles
	c.remainingCycles--
}

func (c *Cpu) init() error {

	// init classic game-boy
	c.setup(CGB)
	c.stopChan = make(chan bool, 1)
	c.doneChan = make(chan bool)

	var opcodes Opcodes
	if err := json.Unmarshal(opcodesFile, &opcodes); err != nil {
		return err
	}
	c.opcodes = &opcodes
	return nil
}

func (c *Cpu) shouldStep() bool {

	if c.step {
		return true
	}

	bg := strings.TrimSpace(c.breakPoints)
	if !c.step && bg == "" {
		return false
	}

	for _, instruction := range strings.Split(bg, " ") {

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
			match := c.pc == Word(n)
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
