package emulator

type mbc1 struct {
	ramSupport     bool
	batterySupport bool
	ramEnabled     bool
	romSelected    uint8
	ramSelected    uint8
	mode           uint8
	ramArea        [32000]uint8
}

func (b *mbc1) Write(area memoryArea, address Word, value uint8) {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#0000---1fff-enable-ram
	if enableRAMArea(address) {
		if value&ENABLE_RAM_MASK == ENABLE_RAM_VALUE {
			b.ramEnabled = true
		} else {
			b.ramEnabled = false
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#2000---3fff-rom-bank
	if b.ramEnabled && selectROMArea(address) {

		// 0x00 -> 0x01 translation
		if value == 0x0 {
			value = 0x1
		}

		switch romSize(area) {
		case 128:
		case 64:
		case 32:
			b.romSelected = value & 0x1F
		case 16:
			b.romSelected = value & 0xF
		case 8:
			b.romSelected = value & 0x7
		case 4:
			b.romSelected = value & 0x3
		case 2:
		default:
			b.romSelected = value & 0x1
		}
	}

	// https://gbdev.io/pandocs/MBC1.html#40005fff--ram-bank-number--or--upper-bits-of-rom-bank-number-write-only
	if selectRAMArea(address) {

		// lower 2 bits of the written value
		b.ramSelected = value & 0x3
	}

	// https://gbdev.io/pandocs/MBC1.html#60007fff--banking-mode-select-write-only
	if selectBankingModeArea(address) {

		if ramSize(area) > 1 && romSize(area) > 32 {
			b.mode = value & 0x1
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#a000---bfff-external-ram
	if b.ramEnabled && externalRAMArea(address) {
		rSize := ramSize(area)

		if rSize == 0 || rSize == 1 {
			b.ramArea[(address-Word(0xA000))%Word(rSize)] = value
		}
		if rSize == 4 {
			if b.mode == 0x1 {
				b.ramArea[(Word(0x2000)*Word(b.ramSelected) + (address - Word(0xA000)))] = value
			} else {
				b.ramArea[address-0xA000] = value
			}
		}
	}
}

func (b *mbc1) Read(area memoryArea, address Word) uint8 {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#0000---3fff-rom-bank-0
	if romBank00(address) {
		if b.mode == 0x0 || romSize(area) <= 32 {
			return area[address]
		} else {
			zeroBankNumber := uint8(0x0)
			if romSize(area) == 64 {
				zeroBankNumber |= (b.ramSelected & 0x1) << 5
			}
			if romSize(area) == 128 {
				zeroBankNumber |= (b.ramSelected & 0x3) << 6
			}
			return area[Word(0x4000)*Word(zeroBankNumber)+address]
		}
	}

	if romBankNN(address) {
		highBankNumber := b.romSelected
		if romSize(area) == 64 {
			highBankNumber |= (b.ramSelected & 0x1) << 5
		}
		if romSize(area) == 128 {
			highBankNumber |= (b.ramSelected & 0x3) << 6
		}
		return area[Word(0x4000)*Word(highBankNumber)+(address-Word(0x4000))]
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#a000---bfff-external-ram
	if b.ramEnabled && externalRAMArea(address) {
		rSize := ramSize(area)

		if rSize == 0 || rSize == 1 {
			return b.ramArea[(address-Word(0xA000))%Word(rSize)]
		}
		if rSize == 4 {
			if b.mode == 0x1 {
				return b.ramArea[(Word(0x2000)*Word(b.ramSelected) + (address - Word(0xA000)))]
			} else {
				return b.ramArea[address-0xA000]
			}
		}
	}

	// open bus value
	return 0xFF
}
