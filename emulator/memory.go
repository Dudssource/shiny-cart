package emulator

// https://gbdev.io/pandocs/Memory_Map.html#memory-map
const (
	ROM_BANK_00_START = Word(0x0000)
	ROM_BANK_00_END   = Word(0x3FFF)

	ROM_BANK_NN_START = Word(0x4000)
	ROM_BANK_NN_END   = Word(0x7FFF)

	VRAM_START = Word(0x8000)
	VRAM_END   = Word(0x9FFF)
)

const (
	SERIAL_TRANSFER_SB = 0xFF01
	SERIAL_TRANSFER_SC = 0xFF02
)

type memoryArea []uint8

type Memory struct {
	mem memoryArea // 8-bit address bus, 64kb memory
	rom memoryArea // ROM area
	mbc *Mbc
}

func NewMemory() *Memory {
	return &Memory{
		mbc: NewMbc(),
	}
}

func (m *Memory) Read(address Word) uint8 {

	// intercept ROM and RAM memory reads
	if m.mbc.initialized() && ownedByMBC(address) {
		return m.mbc.controller.Read(m.rom, address)
	}

	return m.mem[address]
}

func (m *Memory) Write(address Word, value uint8) {

	// intercepts ROM and RAM memory writes
	if m.mbc.initialized() && ownedByMBC(address) {
		m.mbc.controller.Write(m.rom, address, value)
		return
	}

	m.mem[address] = value
}

func (m *Memory) init() error {
	return m.mbc.detectType(m.mem)
}
