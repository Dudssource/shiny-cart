package emulator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_op_jr_imm8(t *testing.T) {
	type args struct {
		c func() *Cpu
	}
	tests := []struct {
		name string
		args args
		want func(t *testing.T, c *Cpu)
	}{
		{
			name: "jump backward",
			args: args{
				c: func() *Cpu {
					cpu := &Cpu{}
					cpu.init(CGB)
					cpu.memory.Write(0x9, 0b00011000) // JR e
					cpu.memory.Write(0xA, 0b11111110) // e=offset -2
					cpu.pc = 0x9                      // PC=0xA, JR e
					fmt.Printf("PC=%X\n", cpu.pc)
					cpu.fetch() // RESULT=0x18 , PC=0xA
					fmt.Printf("PC=%X\n", cpu.pc)
					return cpu
				},
			},
			want: func(t *testing.T, c *Cpu) {
				assert.Equal(t, Word(0x9), c.pc)
			},
		},
		{
			name: "jump forward",
			args: args{
				c: func() *Cpu {
					cpu := &Cpu{}
					cpu.init(CGB)
					cpu.memory.Write(0x9, 0b00011000) // JR e
					cpu.memory.Write(0xA, 0b10)       // e=offset 2
					cpu.pc = 0x9                      // PC=0xA, JR e
					fmt.Printf("PC=%X\n", cpu.pc)
					cpu.fetch() // RESULT=0x18 , PC=0xA
					fmt.Printf("PC=%X\n", cpu.pc)
					return cpu
				},
			},
			want: func(t *testing.T, c *Cpu) {
				assert.Equal(t, Word(0b1101), c.pc)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := tt.args.c()
			op_jr_imm8(cpu, 0x0)
			tt.want(t, cpu)
		})
	}
}
