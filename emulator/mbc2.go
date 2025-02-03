package emulator

type mbc2 struct {
	batterySupport bool
	ramEnabled     bool
	romSelected    uint8
	ramArea        [32000]uint8
	name           string
}

func (b *mbc2) Name() string {
	return b.name
}

func (b *mbc2) Tick() {

}

func (b *mbc2) Write(area memoryArea, address Word, value uint8) {

	//https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc2/index.md#0000---3fff-enable-ram--rom-bank-number
	if address >= ENABLE_RAM_AREA_START && address <= SELECT_ROM_AREA_END {
		if address&0x100 == 0 {
			if value&ENABLE_RAM_MASK == ENABLE_RAM_VALUE {
				//log.Println("Enabling RAM")
				b.ramEnabled = true
			} else {
				//log.Println("Disabling RAM")
				b.ramEnabled = false
			}
		} else {

			// 0x00 -> 0x01 translation
			if value == 0x0 {
				value = 0x1
			}

			b.romSelected = value
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc2/index.md#a000---bfff-external-ram
	if b.ramEnabled && address >= RAM_BANK_START && address <= RAM_BANK_END {
		rAddr := address & 0x1FF
		b.ramArea[rAddr] = value & 0xF
		//log.Printf("Mode 0, written %.8X to RAM bank address %.8X\n", value, rAddr)
	}
}

func (b *mbc2) Read(area memoryArea, address Word) uint8 {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc2/index.md#0000---3fff-rom-bank-0
	if address >= ENABLE_RAM_AREA_START && address <= SELECT_ROM_AREA_END {
		rValue := area[address]
		//log.Printf("Mode 0, read %.8X from ROM bank 00 address %.8X\n", rValue, address)
		return rValue
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc2/index.md#4000---7fff-rom-banks-0x1---0xf
	if address >= SELECT_RAM_AREA_START && address <= SELECT_BANK_MODE_AREA_END {
		bank := Word(b.romSelected)
		rAddr := (bank * SELECT_RAM_AREA_START) + (address - SELECT_RAM_AREA_START)
		rValue := area[rAddr]
		//log.Printf("Read %.8X from ROM bank NN %d address %.8X\n", rValue, bank, rAddr)
		return rValue
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc2/index.md#a000---bfff-external-ram-1
	if b.ramEnabled && address >= RAM_BANK_START && address <= RAM_BANK_END {
		rAddr := address & 0x1FF
		rValue := b.ramArea[rAddr] & 0xF
		//log.Printf("Mode 0 read %.8X from RAM bank 00 address %.8X\n", rValue, rAddr)
		return rValue
	}

	// open bus value
	return 0xFF
}
