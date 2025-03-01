package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpAdd(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		op_add_sp_imm8(c, 0)
		assert.Equal(t, Word(0xFFFA), c.sp)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		op_add_sp_imm8(c, 0)
		assert.Equal(t, Word(0xFFF6), c.sp)
	})

	t.Run("test 3", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w16(reg_hl, 0x8A23)
		c.reg.w16(reg_bc, 0x0605)

		// ADD HL, BC
		op_add_hl_r16(c, 0b00001001)
		assert.Equal(t, Word(0x9028), c.reg.r16(reg_hl))
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&c_flag == 0)
	})

	t.Run("test 4", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w16(reg_hl, 0x8A23)
		c.reg.w16(reg_bc, 0x0605)

		// ADD HL, BC
		op_add_hl_r16(c, 0b00101001)
		assert.Equal(t, Word(0x1446), c.reg.r16(reg_hl))
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
	})

	t.Run("test 5", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x3A)
		c.reg.w8(reg_b, 0xC6)

		// ADD A, B
		op_add_a_r8(c, 0b10000000)
		assert.Equal(t, uint8(0x0), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag > 0)
	})

	t.Run("test 6", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0xE1)
		c.reg.w8(reg_e, 0x0F)
		c.reg.w_flag(c_flag)

		// ADC A, E
		op_adc_a_r8(c, 0b10001011)
		assert.Equal(t, uint8(0xF1), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag == 0)
	})

	t.Run("test 7", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x3B,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0xE1)
		c.reg.w8(reg_e, 0x0F)
		c.reg.w_flag(c_flag)

		// ADC A, 0x3B
		op_adc_a_imm8(c, 0b11001110)
		assert.Equal(t, uint8(0x1D), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
	})

	t.Run("test 8", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x3B,
					0x1E,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0xE1)
		c.reg.w16(reg_hl, 0x1)
		c.reg.w_flag(c_flag)

		// ADC A, (HL)
		op_adc_a_r8(c, 0b10001110)
		assert.Equal(t, uint8(0x00), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
	})
}
