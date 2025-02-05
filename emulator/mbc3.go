package emulator

const (
	RTC_S  = uint8(0x8)
	RTC_M  = uint8(0x9)
	RTC_H  = uint8(0xA)
	RTC_DL = uint8(0xB)
	RTC_DH = uint8(0xC)
)

type mbc3 struct {
	ramSupport        bool
	batterySupport    bool
	rtcSupport        bool
	ramEnabled        bool
	romSelected       uint8 // 7 bit unsigned
	ramSelected       uint8 // 2 bit unsigned
	rtcSelected       uint8
	rtcRegisters      map[uint8]uint8
	ramArea           [32000]uint8
	name              string
	latch             uint8
	remaining         int // ms remaining to increase S register
	rtcRegistersLatch map[uint8]uint8
	mode              uint8
	halted            bool
}

func (b *mbc3) Tick() {
	if b.halted {
		return
	}

	b.remaining++
	if b.remaining < 1000 {
		return
	}
	b.remaining = 0
	b.rtcRegisters[RTC_S] += 1
	if b.rtcRegisters[RTC_S] == 60 {
		b.rtcRegisters[RTC_S] = 0
		b.rtcRegisters[RTC_M] += 1
	}
	if b.rtcRegisters[RTC_M] == 60 {
		b.rtcRegisters[RTC_M] = 0
		b.rtcRegisters[RTC_H] += 1
	}
	if b.rtcRegisters[RTC_H] == 24 {
		b.rtcRegisters[RTC_H] = 0
		// 9 bit day counter + 1
		dc := uint16(b.rtcRegisters[RTC_DL]) | (uint16(b.rtcRegisters[RTC_DH]&0x1) << 8) + 1
		if dc > 0x1FF {
			b.rtcRegisters[RTC_DH] |= 0x80 // set carry flag
		} else {
			b.rtcRegisters[RTC_DL] = uint8(dc & 0xFF)
			b.rtcRegisters[RTC_DH] |= uint8(dc & 0x100 >> 8) // bit 8 of day counter
		}
	}
}

func (b *mbc3) Name() string {
	return b.name
}

func (b *mbc3) Write(area memoryArea, address Word, value uint8) {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc3/index.md#0000---1fff-enable-ram--timer-registers
	if enableRAMArea(address) {
		if value&ENABLE_RAM_MASK == ENABLE_RAM_VALUE {
			//log.Println("Enabling RAM")
			b.ramEnabled = true
		} else {
			//log.Println("Disabling RAM")
			b.ramEnabled = false
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc3/index.md#2000---3fff-rom-bank-low
	if selectROMArea(address) {

		// lower 7 bits of the written value
		value &= 0x7F

		// 0x00 -> 0x01 translation
		if value == 0x0 {
			value = 0x1
		}

		// wrap around
		rs := romSize(area)
		b.romSelected = (value - uint8(rs)) % uint8(rs)

		//log.Printf("Selected ROM bank number %d size=(%d)\n", b.romSelected, romSize(area))
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc3/index.md#4000---5fff-ram-bank--rtc-register-select
	if b.ramEnabled && selectRAMArea(address) {
		if value <= 0x03 {
			// lower 2 bits of the written value
			b.ramSelected = value & 0x3
			b.mode = 0x1
			//log.Printf("Selected RAM bank number %d\n", b.ramSelected)
		} else if b.rtcSupport && value >= 0x08 && value <= 0x0C {
			b.rtcSelected = value
			b.mode = 0x2
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc3/index.md#6000---7fff-rtc-data-latch
	if address >= Word(0x6000) && address <= Word(0x7FFF) {
		if b.latch == 0x0 && value == 0x0 {
			b.latch = 0x1
		}

		if b.latch == 0x1 && value == 0x1 {
			b.latch = 0x0
			for k, v := range b.rtcRegisters {
				b.rtcRegistersLatch[k] = v
			}
		}
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc1/index.md#a000---bfff-external-ram
	if externalRAMArea(address) {
		if b.mode == 0x1 && b.ramEnabled {
			rAddr := (SELECT_ROM_AREA_START*Word(b.ramSelected) + (address - RAM_BANK_START))
			b.ramArea[rAddr] = value
			//log.Printf("Mode 1, written %.8X to RAM bank %d address %.8X\n", value, b.ramSelected, rAddr)
		}

		if b.rtcSupport && b.mode == 0x2 {
			switch b.rtcSelected {
			case RTC_S, RTC_M:
				if b.rtcSelected == RTC_S {
					b.remaining = 0
				}
				b.rtcRegisters[b.rtcSelected] = value & 0x3F
			case RTC_H:
				b.rtcRegisters[b.rtcSelected] = value & 0x1F
			case RTC_DH:
				b.rtcRegisters[b.rtcSelected] = value & 0xC1
				if value&0x40 > 0 {
					b.halted = true
				} else {
					b.halted = false
				}
			default:
				b.rtcRegisters[b.rtcSelected] = value
			}
			if len(b.rtcRegistersLatch) > 0 {
				b.rtcRegistersLatch[b.rtcSelected] = b.rtcRegisters[b.rtcSelected]
			}
		}
	}
}

func (b *mbc3) Read(area memoryArea, address Word) uint8 {

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc3/index.md#0000---3fff-rom-bank-0
	if romBank00(address) {
		rValue := area[address]
		//log.Printf("Mode 1, read %.8X from ROM bank 00 %d address %.8X\n", rValue, bank, rAddr)
		return rValue
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc3/index.md#4000---7fff-rom-banks-0x00---0x7f
	if romBankNN(address) {
		bank := int(b.romSelected)
		rAddr := (bank * SELECT_RAM_AREA_START) + (int(address) - SELECT_RAM_AREA_START)
		rValue := area[rAddr]
		//log.Printf("Read %.8X from ROM bank NN %d address %.8X\n", rValue, bank, rAddr)
		return rValue
	}

	// https://github.com/Hacktix/GBEDG/blob/master/mbcs/mbc3/index.md#a000---bfff-external-ram--rtc-register-1
	if externalRAMArea(address) {
		if b.ramEnabled && b.mode == 0x1 {
			bank := b.ramSelected
			rAddr := (SELECT_ROM_AREA_START*int(bank) + (int(address) - RAM_BANK_START))
			rValue := b.ramArea[rAddr]
			//log.Printf("Mode 0 read %.8X from RAM bank 00 address %.8X\n", rValue, rAddr)
			return rValue
		}

		if b.rtcSupport && b.mode == 0x2 {
			if len(b.rtcRegistersLatch) == 0 {
				// no latch
				return 0xFF
			}

			rVal := b.rtcRegistersLatch[b.rtcSelected]
			switch b.rtcSelected {
			case RTC_S, RTC_M:
				return rVal | 0xC0
			case RTC_H:
				return rVal | 0xE0
			case RTC_DH:
				return rVal | 0x3E
			default:
				return rVal
			}
		}
	}

	// open bus value
	return 0xFF
}
