package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LDH__C_,A
func op_ldh_c_a(c *Cpu, _ uint8) {
	c.requiredCycles = 2
	c.memory.Write(Word(0xFF00+uint16(c.reg.r8(reg_c))), c.reg.r8(reg_a))
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LDH__n16_,A
func op_ldh_imm8_a(c *Cpu, _ uint8) {
	c.requiredCycles = 3
	z := c.fetch()
	c.memory.Write(NewWord(0xFF, z), c.reg.r8(reg_a))
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LDH_A,_C_
func op_ldh_a_c(c *Cpu, _ uint8) {
	c.requiredCycles = 2
	c.reg.w8(reg_a, c.memory.Read(Word(0xFF00+uint16(c.reg.r8(reg_c)))))
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LDH_A,_n16_
func op_ldh_a_imm8(c *Cpu, _ uint8) {
	c.requiredCycles = 3
	z := c.fetch()
	z1 := c.memory.Read(NewWord(0xFF, z))
	c.reg.w8(reg_a, z1)

}
