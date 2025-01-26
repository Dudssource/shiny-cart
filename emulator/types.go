package emulator

type Memory struct {
	mem [65536]uint8 // 8-bit address bus, 64kb memory
}

func (m *Memory) Read(address Word) uint8 {
	return m.mem[address]
}

func (m *Memory) Write(address Word, value uint8) {
	m.mem[address] = value
}

type Word uint16

func NewWord(hi uint8, lo uint8) Word {
	return Word(uint16(hi)<<8 | uint16(lo))
}

func (w Word) High() uint8 {
	return uint8(w & 0xFF00 >> 8)
}

func (w Word) Low() uint8 {
	return uint8(w & 0xFF)
}

func SetLow(w Word, low uint8) Word {
	return (w & 0xFF00) | Word(low)
}

func SetHigh(w Word, high uint8) Word {
	return Word(high)<<8 | (w & 0xFF)
}

type instruction func(c *Cpu, opcode uint8)
