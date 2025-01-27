package emulator

import "fmt"

const (
	// https://gbdev.io/pandocs/The_Cartridge_Header.html#0147--cartridge-type
	CARTRIDGE_HEADER_TYPE     = 0x0147
	CARTRIDGE_HEADER_ROM_SIZE = 0x0148
	CARTRIDGE_HEADER_RAM_SIZE = 0x0149
)

// https://gbdev.io/pandocs/The_Cartridge_Header.html#0147--cartridge-type
// https://github.com/Hacktix/GBEDG/blob/master/mbcs/index.md#how-to-detect-a-roms-mbc
const (
	ROM_ONLY        = 0x00
	ROM_RAM         = 0x08
	ROM_RAM_BATTERY = 0x09

	MBC1             = 0x01
	MBC1_RAM         = 0x02
	MBC1_RAM_BATTERY = 0x03

	MBC2         = 0x05
	MBC2_BATTERY = 0x06

	MBC3                   = 0x11
	MBC3_TIMER_BATTERY     = 0x0F
	MBC3_TIMER_RAM_BATTERY = 0x10

	MBC5                    = 0x19
	MBC5_RAM                = 0x1A
	MBC5_RAM_BATTERY        = 0x1B
	MBC5_RUMBLE             = 0x1C
	MBC5_RUMBLE_RAM         = 0x1D
	MBC5_RUMBLE_RAM_BATTERY = 0x1E

	MBC6 = 0x20
)

const (
	/* REGISTERS */
	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#0000---1fff-enable-ram
	ENABLE_RAM_AREA_START = Word(0x0000)
	ENABLE_RAM_AREA_END   = Word(0x1FFF)
	ENABLE_RAM_VALUE      = uint8(0xA)
	ENABLE_RAM_MASK       = uint8(0xF)

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#2000---3fff-rom-bank
	SELECT_ROM_AREA_START = Word(0x2000)
	SELECT_ROM_AREA_END   = Word(0x3FFF)

	// https://gbdev.io/pandocs/MBC1.html#40005fff--ram-bank-number--or--upper-bits-of-rom-bank-number-write-only
	SELECT_RAM_AREA_START = Word(0x4000)
	SELECT_RAM_AREA_END   = Word(0x5FFF)

	// https://gbdev.io/pandocs/MBC1.html#a000bfff--ram-bank-0003-if-any
	RAM_BANK_START = Word(0xA000)
	RAM_BANK_END   = Word(0xBFFF)

	// https://gbdev.io/pandocs/MBC1.html#60007fff--banking-mode-select-write-only
	SELECT_BANK_MODE_AREA_START = Word(0x6000)
	SELECT_BANK_MODE_AREA_END   = Word(0x7FFF)
)

var (

	// romSizeMap ROM size in banks
	// https://gbdev.io/pandocs/The_Cartridge_Header.html#0148--rom-size
	romSizeMap = map[uint8]int{
		0x00: 2,   // no banking / 32 Kib
		0x01: 4,   // 64 Kib
		0x02: 8,   // 128 Kib
		0x03: 16,  // 256 Kib
		0x04: 32,  // 512 Kib
		0x05: 64,  // 1 Mib
		0x06: 128, // 2 Mib
		0x07: 256, // 4 Mib
		0x08: 512, // 8 Mib
		0x52: 72,  // 1.1 Mib
		0x53: 80,  // 1.2 Mib
		0x54: 96,  // 1.5 Mib
	}

	// ramSizeMap RAM size in banks
	// https://gbdev.io/pandocs/The_Cartridge_Header.html#0149--ram-size
	ramSizeMap = map[uint8]int{
		0x00: 0,  // No RAM
		0x01: 0,  // Unused / 2 Kib
		0x02: 1,  // 8 Kib
		0x03: 4,  // 32 Kib
		0x04: 16, // 128 Kib
		0x05: 8,  // 64 Kib
	}

	mbcControllerMap = map[int]memoryController{
		MBC1:             &mbc1{},
		MBC1_RAM:         &mbc1{ramSupport: true},
		MBC1_RAM_BATTERY: &mbc1{ramSupport: true, batterySupport: true},
	}
)

type memoryController interface {
	Write(area memoryArea, address Word, value uint8)
	Read(area memoryArea, address Word) uint8
}

type Mbc struct {
	controller memoryController
	mem        memoryArea
}

func NewMbc() *Mbc {
	return &Mbc{}
}

func (m *Mbc) detectType(mem memoryArea) error {

	cartridgeType := mem[CARTRIDGE_HEADER_TYPE]

	// no MBC
	if cartridgeType == 0x0 {
		return nil
	}

	controller, ok := mbcControllerMap[int(cartridgeType)]
	if !ok {
		return fmt.Errorf("not supported cartridge type %X", cartridgeType)
	}
	m.controller = controller
	m.mem = mem
	return nil
}

func (m *Mbc) initialized() bool {
	return m.controller != nil && len(m.mem) > 0
}

func ownedByMBC(address Word) bool {
	return (address >= ROM_BANK_00_START && address <= ROM_BANK_NN_END) || (address >= RAM_BANK_START && address <= RAM_BANK_END)
}

func enableRAMArea(address Word) bool {
	return address >= ENABLE_RAM_AREA_START && address <= ENABLE_RAM_AREA_END
}

func selectROMArea(address Word) bool {
	return address >= SELECT_ROM_AREA_START && address <= SELECT_ROM_AREA_END
}

func romBank00(address Word) bool {
	return address >= ROM_BANK_00_START && address <= ROM_BANK_00_END
}

func romBankNN(address Word) bool {
	return address >= ROM_BANK_NN_START && address <= ROM_BANK_NN_END
}

func selectRAMArea(address Word) bool {
	return address >= SELECT_RAM_AREA_START && address <= SELECT_RAM_AREA_END
}

func externalRAMArea(address Word) bool {
	return address >= RAM_BANK_START && address <= RAM_BANK_END
}

func selectBankingModeArea(address Word) bool {
	return address >= SELECT_BANK_MODE_AREA_START && address <= SELECT_BANK_MODE_AREA_END
}

func romSize(area memoryArea) int {
	return romSizeMap[area[CARTRIDGE_HEADER_ROM_SIZE]]
}

// ramSize in KiB
func ramSize(area memoryArea) int {
	return ramSizeMap[area[CARTRIDGE_HEADER_RAM_SIZE]]
}
