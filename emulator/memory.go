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
	PORT_SERIAL_TRANSFER_SB = 0xFF01
	PORT_SERIAL_TRANSFER_SC = 0xFF02
	PORT_OAM_DMA_CONTROL    = 0xFF46
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
	if m.mbc != nil && m.mbc.initialized() && ownedByMBC(address) {
		return m.mbc.controller.Read(m.rom, address)
	}

	rVal := m.mem[address]

	// unreadable bits return 1
	if address == PORT_JOYPAD {
		return rVal | 0xC0
	} else if address == PORT_SERIAL_TRANSFER_SC {
		return rVal | 0x7E
	} else if address == PORT_TAC {
		return rVal | 0xF8
	} else if address == INTERRUPT_FLAG {
		return rVal | 0xE0
	} else if address == LCD_REGISTER {
		return rVal | 0x80
	}

	// audio/wave
	if address == 0xFF26 {
		return rVal | 0x70
	} else if address == 0xFF10 {
		return rVal | 0x80
	} else if address == 0xFF14 {
		return rVal | 0x38
	} else if address == 0xFF1A {
		return rVal | 0x7F
	} else if address == 0xFF1C {
		return rVal | 0x9F
	} else if address == 0xFF1E {
		return rVal | 0x38
	} else if address == 0xFF20 {
		return rVal | 0xC0
	} else if address == 0xFF23 {
		return rVal | 0x3F
	}

	// CGB
	if address == 0xFF4F || address == 0xFF68 || address == 0xFF6A || address == 0xFF72 || address == 0xFF73 ||
		address == 0xFF74 || address == 0xFF75 || address == 0xFF76 || address == 0xFF77 {
		return rVal | 0xFF
	}

	// unmapped
	if address == 0xFF03 || address == 0xFF08 || address == 0xFF09 || address == 0xFF0A || address == 0xFF0B || address == 0xFF0C ||
		address == 0xFF0D || address == 0xFF0E || address == 0xFF15 || address == 0xFF1F || address == 0xFF27 || address == 0xFF28 ||
		address == 0xFF29 || address == 0xFF4C || address == 0xFF4D || address == 0xFF4E || address == 0xFF69 || address == 0xFF74 {
		return 0xFF
	}

	// unmapped 2
	if (address >= 0xFF6B && address <= 0xFF71) || (address >= 0xFF50 && address <= 0xFF67) || (address >= 0xFF78 && address <= 0xFF7F) {
		return 0xFF
	}

	return rVal
}

func (m *Memory) Write(address Word, value uint8) {

	// intercepts ROM and RAM memory writes
	if m.mbc != nil && m.mbc.initialized() && ownedByMBC(address) {
		m.mbc.controller.Write(m.rom, address, value)
		return
	}

	m.mem[address] = value
}

func (m *Memory) init() error {
	return m.mbc.detectType(m.mem)
}
