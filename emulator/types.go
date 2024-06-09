package emulator

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

type InstructionSet struct {
	execute instruction
	cycles  int
}

type instruction func(c *Cpu, opcode uint8)
