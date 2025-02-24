package emulator

import (
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
					cpu.init()
					cpu.memory.Write(0x9, 0x18) // JR e
					cpu.memory.Write(0xA, 0xFE) // e=offset -2
					cpu.pc = 0x9                // PC=0xA, JR e
					cpu.fetch()                 // RESULT=0x18 , PC=0xA
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
					cpu.init()
					cpu.memory.Write(0x9, 0x18) // JR e
					cpu.memory.Write(0xA, 0x2)  // e=offset 2
					cpu.pc = 0x9                // PC=0xA, JR e
					cpu.fetch()                 // RESULT=0x18 , PC=0xA
					return cpu
				},
			},
			want: func(t *testing.T, c *Cpu) {
				assert.Equal(t, Word(0xD), c.pc)
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
