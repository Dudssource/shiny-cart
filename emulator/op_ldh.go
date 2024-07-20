package emulator

func op_ldh_c_a(c *Cpu, opcode uint8) {
	c.requiredCycles = 2
}
