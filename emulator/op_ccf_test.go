package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_op_ccf(t *testing.T) {
	type args struct {
		c func() *Cpu
	}
	tests := []struct {
		name string
		args args
		want func(t *testing.T, c *Cpu)
	}{
		{
			name: "set carry off",
			args: args{
				c: func() *Cpu {
					cpu := &Cpu{}
					cpu.init(CGB)
					cpu.reg.w_flag(c_flag)
					return cpu
				},
			},
			want: func(t *testing.T, c *Cpu) {
				assert.Equal(t, flag(0x0), c.reg.r_flags())
			},
		},
		{
			name: "set carry on",
			args: args{
				c: func() *Cpu {
					cpu := &Cpu{}
					cpu.init(CGB)
					cpu.reg.w_flag(0x0)
					return cpu
				},
			},
			want: func(t *testing.T, c *Cpu) {
				assert.Equal(t, c_flag, c.reg.r_flags())
			},
		},
		{
			name: "set carry off keeping",
			args: args{
				c: func() *Cpu {
					cpu := &Cpu{}
					cpu.init(CGB)
					cpu.reg.w_flag(n_flag | h_flag | z_flag | c_flag)
					return cpu
				},
			},
			want: func(t *testing.T, c *Cpu) {
				assert.Equal(t, n_flag|h_flag|z_flag, c.reg.r_flags())
			},
		},
		{
			name: "set carry on keeping",
			args: args{
				c: func() *Cpu {
					cpu := &Cpu{}
					cpu.init(CGB)
					cpu.reg.w_flag(n_flag | h_flag | z_flag)
					return cpu
				},
			},
			want: func(t *testing.T, c *Cpu) {
				assert.Equal(t, n_flag|h_flag|z_flag|c_flag, c.reg.r_flags())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := tt.args.c()
			op_ccf(cpu, 0x0)
			tt.want(t, cpu)
		})
	}
}
