package emulator

type mbc5 struct {
	ramSupport     bool
	rumbleSupport  bool
	batterySupport bool
	ramEnabled     bool
	romSelected    uint16 // 9-bit unsigned
	ramSelected    uint8  // 4 bit unsigned
	ramArea        [32768]uint8
	name           string
}

func (b *mbc5) Name() string {
	return b.name
}

func (b *mbc5) Tick() {

}

func (b *mbc5) Write(area memoryArea, address Word, value uint8) {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#0000---1fff-enable-ram
	if enableRAMArea(address) {
		if value&ENABLE_RAM_MASK == ENABLE_RAM_VALUE {
			//log.Println("Enabling RAM")
			b.ramEnabled = true
		} else {
			//log.Println("Disabling RAM")
			b.ramEnabled = false
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#2000---2fff-rom-bank-low
	if address >= SELECT_ROM_AREA_START && address <= Word(0x2FFF) {
		b.romSelected = (b.romSelected & 0xFF00) | uint16(value)
		// wrap around
		rs := romSize(area)
		b.romSelected = (b.romSelected - uint16(rs)) % uint16(rs)
		//log.Printf("Selected ROM bank number %d size=%d, src=%b\n", b.romSelected, romSize(area), value)
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#3000---3fff-rom-bank-high
	if address >= Word(0x3000) && address <= SELECT_ROM_AREA_END {
		b.romSelected = (b.romSelected & 0xFF) | ((uint16(value) & 0x1) << 8)
		// wrap around
		rs := romSize(area)
		b.romSelected = (b.romSelected - uint16(rs)) % uint16(rs)
		//log.Printf("Selected ROM bank number %d size=%d, src=%b\n", b.romSelected, romSize(area), value)
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#4000---5fff-ram-bank
	if b.ramEnabled && selectRAMArea(address) {

		if !b.rumbleSupport {
			// lower 4 bits of the written value
			b.ramSelected = value & 0xF
		} else {
			b.ramSelected = value & 0x7
		}
		//log.Printf("Selected RAM bank number %d\n", b.ramSelected)
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#a000---bfff-external-ram
	if b.ramEnabled && externalRAMArea(address) {
		rAddr := (0x2000 * int(b.ramSelected)) + (int(address) - 0xA000)
		b.ramArea[rAddr] = value
		//log.Printf("Mode 1, 	written %.8X to RAM bank %d address %.8X\n", value, b.ramSelected, rAddr)
	}
}

func (b *mbc5) Read(area memoryArea, address Word) uint8 {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#0000---3fff-rom-bank-0
	if romBank00(address) {
		rValue := area[address]
		//log.Printf("Mode 0, read %.8X from ROM bank 00 address %.8X\n", rValue, address)
		return rValue
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#4000---7fff-rom-banks-0x000---0x1ff
	if romBankNN(address) {
		bank := int(b.romSelected)
		rAddr := (bank * SELECT_RAM_AREA_START) + (int(address) - SELECT_RAM_AREA_START)
		rValue := area[rAddr]
		//log.Printf("Read %.8X from ROM bank NN %d address %.8X\n", rValue, bank, rAddr)
		return rValue
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc5/index.md#a000---bfff-external-ram-1
	if b.ramEnabled && externalRAMArea(address) {
		bank := b.ramSelected
		rAddr := (SELECT_ROM_AREA_START*int(bank) + (int(address) - RAM_BANK_START))
		rValue := b.ramArea[rAddr]
		//log.Printf("Mode 1, read %.8X from RAM bank NN %d address %.8X\n", rValue, bank, rAddr)
		return rValue
	}

	// open bus value
	return 0xFF
}
