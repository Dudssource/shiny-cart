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
