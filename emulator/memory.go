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
	mem        memoryArea // 8-bit address bus, 64kb memory
	rom        memoryArea // ROM area
	mbc        *Mbc
	joypad     uint8
	dma        bool
	resetTimer bool
	sound      *Sound
}

func NewMemory(sound *Sound, mem memoryArea) *Memory {
	return &Memory{
		mbc:    NewMbc(),
		joypad: 0xFF,
		sound:  sound,
		mem:    mem,
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
		rVal |= 0xCF
		state := m.joypad
		if rVal&0x20 == 0 {
			rVal = (rVal & 0xF0) | ((state & 0xF0) >> 4)
		}
		if rVal&0x10 == 0 {
			rVal = (rVal & 0xF0) | (state & 0xF)
		}
		return rVal
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

	if address == PORT_DIV {
		// reset timer
		m.mem[address] = 0x0
		// reset tima
		m.mem[PORT_TIMA] = 0x0

		// timer v2
		m.resetTimer = true
		return
	}

	// APU off all registers are read-only
	if m.mem[NR52]&0x80 > 0 {

		// trigger event SCH1
		if address == NR14 && (value&0x80) > 0 {
			// reset SCH1
			m.sound.resetSC1()
		}

		// trigger event SCH2
		if address == NR24 && (value&0x80) > 0 {
			// reset SCH2
			m.sound.resetSC2()
		}

		// trigger event SCH3
		if address == NR34 && (value&0x80) > 0 {
			// reset SCH3
			m.sound.resetSC3()
		}
	}

	// Length counter SCH1
	if address == NR11 {
		m.sound.lengthCounterSC1 = 64 - int(value&0x3F)
	}

	// Length counter SCH2
	if address == NR21 {
		m.sound.lengthCounterSC2 = 64 - int(value&0x3F)
	}

	// Length counter SCH3
	if address == NR31 {
		m.sound.lengthCounterSC3 = 256 - int(value)
	}

	// Length counter SCH4
	if address == NR41 {
		m.sound.lengthCounterSC4 = 64 - int(value&0x3F)
	}

	// turn APU off
	if address == NR52 && (value&0x80) == 0 {
		m.sound.powerOff()
	}

	if address == PORT_JOYPAD {
		// write JP
		m.mem[address] = (value & 0x30) | (m.mem[address] & 0xCF)
		return
	}

	if address == PORT_OAM_DMA_CONTROL {
		// write DMA
		m.mem[address] = value
		m.dma = true
		return
	}

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
