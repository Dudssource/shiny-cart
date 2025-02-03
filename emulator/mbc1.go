package emulator

type mbc1 struct {
	ramSupport     bool
	batterySupport bool
	ramEnabled     bool
	romSelected    uint8
	ramSelected    uint8
	mode           uint8
	ramArea        [32000]uint8
	name           string
}

func (b *mbc1) Tick() {

}

func (b *mbc1) Name() string {
	return b.name
}

func (b *mbc1) Write(area memoryArea, address Word, value uint8) {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#0000---1fff-enable-ram
	if enableRAMArea(address) {
		if value&ENABLE_RAM_MASK == ENABLE_RAM_VALUE {
			//log.Println("Enabling RAM")
			b.ramEnabled = true
		} else {
			//log.Println("Disabling RAM")
			b.ramEnabled = false
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#2000---3fff-rom-bank
	if selectROMArea(address) {

		// lower 5 bits of the written value
		value &= 0x1F

		// 0x00 -> 0x01 translation
		if value == 0x0 {
			value = 0x1
		}

		// ignore pins according to the rom size
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

		//log.Printf("Selected ROM bank number %d size=(%d)\n", b.romSelected, romSize(area))
	}

	// https://gbdev.io/pandocs/MBC1.html#40005fff--ram-bank-number--or--upper-bits-of-rom-bank-number-write-only
	if b.ramEnabled && selectRAMArea(address) {

		// lower 2 bits of the written value
		b.ramSelected = value & 0x3

		//log.Printf("Selected RAM bank number %d\n", b.ramSelected)
	}

	// https://gbdev.io/pandocs/MBC1.html#60007fff--banking-mode-select-write-only
	if selectBankingModeArea(address) {
		b.mode = value & 0x1
		//log.Printf("Selected ROM mode %d\n", b.mode)
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#a000---bfff-external-ram
	if b.ramEnabled && externalRAMArea(address) {
		if b.mode == 0x0 {
			rAddr := address - RAM_BANK_START
			b.ramArea[rAddr] = value
			//log.Printf("Mode 0, written %.8X to RAM bank address %.8X\n", value, rAddr)
		} else {
			rAddr := (SELECT_ROM_AREA_START*Word(b.ramSelected) + (address - RAM_BANK_START))
			// TODO: Review
			// if ramSize(area) <= 8 {
			// 	rAddr = (address - RAM_BANK_START) % Word(ramSize(area)*1024)
			// }
			b.ramArea[rAddr] = value
			//log.Printf("Mode 1, written %.8X to RAM bank %d address %.8X\n", value, b.ramSelected, rAddr)
		}
	}
}

func (b *mbc1) Read(area memoryArea, address Word) uint8 {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#0000---3fff-rom-bank-0
	if romBank00(address) {
		if b.mode == 0x0 {
			rValue := area[address]
			//log.Printf("Mode 0, read %.8X from ROM bank 00 address %.8X\n", rValue, address)
			return rValue
		} else {
			bank := Word(0x0)
			if romSize(area) > 32 {
				bank = Word(b.ramSelected << 5)
			}
			rAddr := (bank * SELECT_RAM_AREA_START) + address
			rValue := area[rAddr]
			//log.Printf("Mode 1, read %.8X from ROM bank 00 %d address %.8X\n", rValue, bank, rAddr)
			return rValue
		}
	}

	if romBankNN(address) {
		bank := Word(b.ramSelected<<5 | b.romSelected)
		rAddr := (bank * SELECT_RAM_AREA_START) + (address - SELECT_RAM_AREA_START)
		rValue := area[rAddr]
		//log.Printf("Read %.8X from ROM bank NN %d address %.8X\n", rValue, bank, rAddr)
		return rValue
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#a000---bfff-external-ram
	if b.ramEnabled && externalRAMArea(address) {
		if b.mode == 0x0 {
			rAddr := address - RAM_BANK_START
			rValue := b.ramArea[rAddr]
			//log.Printf("Mode 0 read %.8X from RAM bank 00 address %.8X\n", rValue, rAddr)
			return rValue
		} else {
			bank := b.ramSelected
			rAddr := (SELECT_ROM_AREA_START*Word(bank) + (address - RAM_BANK_START))
			rValue := b.ramArea[rAddr]
			//log.Printf("Mode 1, read %.8X from RAM bank NN %d address %.8X\n", rValue, bank, rAddr)
			return rValue
		}
	}

	// open bus value
	return 0xFF
}
