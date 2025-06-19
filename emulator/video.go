package emulator

import (
	"fmt"
	"image/color"
	"log"
	"sort"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Pixel uint8

type Tile [16][8]Pixel

type Sprite struct {
	yPos  uint8
	xPos  uint8
	tile  uint8
	flags uint8
}

func (t Sprite) String() string {
	return fmt.Sprintf("{y=%d x=%d t=%d f=%.8b}", t.yPos, t.xPos, t.tile, t.flags)
}

const (
	VRAM_BACKGROUND_START = 0x9800
	VRAM_BACKGROUND_END   = 0x9BFF

	VRAM_WINDOW_START = 0x9C00
	VRAM_WINDOW_END   = 0x9FFF

	OAM_MEMORY_START = 0xFE00
	OAM_MEMORY_END   = 0xFE9F

	LCDC_REGISTER = 0xFF40
	LCD_REGISTER  = 0xFF41
	LY_REGISTER   = 0xFF44
	LYC_REGISTER  = 0xFF45
)

var palettes = map[int]map[Pixel]color.RGBA{
	// default b/w
	0: {
		0: rl.RayWhite,
		1: rl.LightGray,
		2: rl.DarkGray,
		3: rl.Black,
	},
	// https://www.deviantart.com/thewolfbunny64/art/Game-Boy-Palette-Lime-Midori-810574708
	1: {
		0: rl.NewColor(224, 235, 175, 255),
		1: rl.NewColor(170, 207, 83, 255),
		2: rl.NewColor(123, 141, 66, 255),
		3: rl.NewColor(71, 89, 80, 255),
	},
	// https://www.deviantart.com/thewolfbunny64/art/Game-Boy-Palette-Pokemon-Pinball-Ver-882658817
	2: {
		0: rl.NewColor(232, 248, 184, 255),
		1: rl.NewColor(160, 176, 80, 255),
		2: rl.NewColor(120, 96, 48, 255),
		3: rl.NewColor(24, 24, 32, 255),
	},
	// https://www.deviantart.com/thewolfbunny64/art/Game-Boy-Palette-Green-Awakening-883049033
	3: {
		0: rl.NewColor(241, 255, 221, 255),
		1: rl.NewColor(152, 219, 117, 255),
		2: rl.NewColor(54, 112, 88, 255),
		3: rl.NewColor(0, 11, 22, 255),
	},
}

type Video struct {
	w              int32
	h              int32
	internalWidth  int32
	internalHeight int32
	scaleFactor    int32

	mem        *Memory
	scanline   int
	scancolumn int

	mode           uint8
	buffer         []Sprite
	tick           int
	delay          int
	nextMode       uint8
	disabled       bool
	lastComparison bool
	palette        int

	currentOamAddr Word

	videoMemory [144][160]Pixel

	ts     int
	t      time.Time
	total  int
	total2 int
}

func (v *Video) init(width, height int32) {
	v.internalWidth = 64
	v.internalHeight = 32
	v.w = width
	v.h = height
	v.scaleFactor = v.w / v.internalWidth
	v.t = time.Now()

	rl.InitWindow(v.w, v.h, "GameBoy-DMG Emulator")
	rl.SetTargetFPS(60)
}

func (v *Video) setMode(mode uint8) {

	if mode != v.mode {
		//fmt.Printf("TOTAL MODE %d %d\n", v.mode, v.total)
		//fmt.Printf("Set mode=%d\n", mode)
		v.total = 0
	}
	lcds := v.mem.Read(LCD_REGISTER)

	if (lcds&0x20 > 0 && mode == 2) || (lcds&0x10 > 0 && mode == 1) || (lcds&0x8 > 0 && mode == 0) {
		iflag := v.mem.Read(INTERRUPT_FLAG) | 0x2
		v.mem.Write(INTERRUPT_FLAG, iflag)
		//v.lastComparison = true
	}

	if mode == 1 {
		// v-blank interrupt
		iflag := v.mem.Read(INTERRUPT_FLAG) | 0x1
		v.mem.Write(INTERRUPT_FLAG, iflag)
		v.ts++
	}

	if time.Since(v.t).Seconds() >= 1 {
		//fmt.Printf("FPS %d\n", v.ts)
		v.ts = 0
		v.t = time.Now()
	}

	// set mode
	v.mode = mode & 0x3
	v.mem.Write(LCD_REGISTER, (v.mem.Read(LCD_REGISTER)&0xFC)|v.mode)
}

func (v *Video) height() uint8 {
	objSize := v.mem.Read(LCDC_REGISTER) & 0x4 >> 2
	if objSize == 0x0 {
		return 8
	}
	return 16
}

func (v *Video) readOAMSprite(addr Word) Sprite {

	sprite := Sprite{}
	sprite.yPos = v.mem.Read(Word(addr+v.currentOamAddr)) - 16
	v.currentOamAddr++
	sprite.xPos = v.mem.Read(Word(addr+v.currentOamAddr)) - 8
	v.currentOamAddr++
	sprite.tile = v.mem.Read(Word(addr + v.currentOamAddr))
	v.currentOamAddr++
	sprite.flags = v.mem.Read(Word(addr + v.currentOamAddr))
	v.currentOamAddr++

	return sprite
}

func (v *Video) shouldAddToBuffer(sprite Sprite) bool {
	return len(v.buffer) < 10 && v.scanline >= int(sprite.yPos) && int(sprite.yPos)+int(v.height()) > v.scanline
}

func (v *Video) fetchWindow() (Pixel, Pixel, bool) {

	lcdc := v.mem.Read(LCDC_REGISTER)

	// bg disabled or window disabled
	if lcdc&0x1 == 0 || lcdc&0x20 == 0 {
		return 0, 0, false
	}

	tileMap := (v.mem.Read(LCDC_REGISTER) & 0x40) >> 6
	address := 0x9800
	if tileMap == 1 {
		address = 0x9C00
	}

	y := v.scanline
	x := v.scancolumn

	wy := int(v.mem.Read(0xFF4A))
	wx := int(v.mem.Read(0xFF4B) - 7)

	// bg disabled or outside window boundaries
	if wy <= y && x >= wx {

		mode := (lcdc & 0x10) >> 4
		yPos := (y - wy)
		xPos := (x - wx)

		offset := ((xPos / 8) + (32 * (int(yPos / 8))))

		tileNumber := v.mem.Read(Word(address + offset))

		var tileDataAddress Word

		if mode == 1 {
			tileDataAddress = (Word(int(tileNumber) * 16)) + 0x8000
		} else {
			// check sign bit
			signed := ((tileNumber & 0x80) >> 7) > 0
			if signed {
				tileDataAddress = Word(0x9000 - ((int(^tileNumber + 1)) * 16))
			} else {
				tileDataAddress = Word(0x9000 + (int(tileNumber) * 16))
			}
		}

		tileDataAddress += Word(2 * (yPos % 8))

		l1 := v.mem.Read(tileDataAddress)
		l2 := v.mem.Read(tileDataAddress + 1)

		var tile [8]Pixel

		for b := 0; b < 8; b++ {
			pixels := ((l2&(0x1<<b))>>b)<<1 | (l1&(0x1<<b))>>b
			tile[7-b] = Pixel(pixels)
		}

		pixel := tile[(xPos)%8]
		palette := v.mem.Read(0xFF47)
		color := palette & ((0x3) << (pixel * 2)) >> (pixel * 2)
		return Pixel(color), pixel, true
	}

	return 0, 0, false
}

// fetchBackground
// https://github.com/Hacktix/GBEDG/blob/master/ppu/index.md#background-pixel-fetching
func (v *Video) fetchBackground() (Pixel, Pixel) {

	lcdc := v.mem.Read(LCDC_REGISTER)

	// bg disabled
	if lcdc&0x1 == 0 {
		return 0, 0
	}

	tileMap := lcdc & 0x8 >> 3
	address := 0x9800
	if tileMap == 1 {
		address = 0x9C00
	}

	y := v.scanline
	x := v.scancolumn

	scy := int(v.mem.Read(0xFF42))
	scx := int(v.mem.Read(0xFF43))

	offset := ((((x + scx) / 8) & 0x1F) + (32 * (((y + scy) & 0xFF) / 8))) & 0x3FF

	mode := lcdc & 0x10 >> 4
	tileNumber := v.mem.Read(Word(address + offset))

	var tileDataAddress Word

	if mode == 1 {
		tileDataAddress = (Word(int(tileNumber) * 16)) + 0x8000
	} else {
		// check sign bit
		signed := ((tileNumber & 0x80) >> 7) > 0
		if signed {
			tileDataAddress = Word(0x9000 - ((int(^tileNumber + 1)) * 16))
		} else {
			tileDataAddress = Word(0x9000 + (int(tileNumber) * 16))
		}
	}

	tileDataAddress += Word(2 * ((y + scy) % 8))

	l1 := v.mem.Read(tileDataAddress)
	l2 := v.mem.Read(tileDataAddress + 1)

	var tile [8]Pixel

	for b := 0; b < 8; b++ {
		pixels := ((l2&(0x1<<b))>>b)<<1 | (l1&(0x1<<b))>>b
		tile[7-b] = Pixel(pixels)
	}

	pixel := tile[(scx+x)%8]
	palette := v.mem.Read(0xFF47)
	color := palette & ((0x3) << (pixel * 2)) >> (pixel * 2)
	return Pixel(color), pixel
}

func (v *Video) fetchTile(tileNumber, mode, size uint8) Tile {
	var tileAddress Word

	if mode == 1 {
		tileAddress = (Word(int(tileNumber) * 16)) + 0x8000
	} else {
		// check sign bit
		signed := ((tileNumber & 0x80) >> 7) > 0
		if signed {
			tileAddress = Word(0x9000 - (int(^tileNumber+1) * 16))
		} else {
			tileAddress = Word(0x9000 + (int(tileNumber) * 16))
		}
	}

	var tile [16][8]Pixel

	// loop to build the pixels, byte per byte
	for s := Word(0x0); s < Word(size); s++ {
		l1 := v.mem.Read(tileAddress)
		tileAddress++
		l2 := v.mem.Read(tileAddress)
		tileAddress++
		for b := 0; b < 8; b++ {
			pixels := ((l2&(0x1<<b))>>b)<<1 | (l1&(0x1<<b))>>b
			tile[s][7-b] = Pixel(pixels)
		}
	}

	return tile
}

func (v *Video) draw() {

	rl.BeginDrawing()
	rl.ClearBackground(palettes[v.palette][0])
	defer rl.EndDrawing()

	colors := palettes[v.palette]

	// Draw
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			var (
				scaleFactory = int32(4)
				posX         = int32(x) * scaleFactory
				posY         = int32(y) * scaleFactory
			)
			pixel := v.videoMemory[y][x]

			color := colors[pixel]
			if pixel == 100 {
				color = rl.ColorAlpha(color, 0)
			}

			rl.DrawRectangle(posX, posY, scaleFactory, scaleFactory, color)
		}
	}
}

func (v *Video) checkInterruption(resetCondition bool) {
	lcds := v.mem.Read(LCD_REGISTER)
	lyc := v.mem.Read(LYC_REGISTER)

	// LYC enabled, LYC=LY CONDITION IS MET, LY=LYC COMPARISON IS FALSE (IRQ BLOCKING)
	if lcds&0x40 > 0 && lyc == uint8(v.scanline) {
		if lcds&0x4 == 0 {
			iflag := v.mem.Read(INTERRUPT_FLAG) | 0x2
			v.mem.Write(INTERRUPT_FLAG, iflag)
			v.mem.Write(LCD_REGISTER, lcds|0x4)
			// log.Printf("REQUESTING INTERRUPTION LYC=%d LY=%d LCDS=%.8b LCDC=%.8b IME=%d PC=0x%X\n", lyc, v.scanline, lcds, v.mem.Read(LCDC_REGISTER), c.ime, c.pc)
		}
		//v.lastComparison = true
	} else if resetCondition {
		// reset LY=LYC COMPARISON FLAG
		v.mem.Write(LCD_REGISTER, lcds&0xFB)
	}
}

func (v *Video) advanceLy(_ *Cpu) {
	v.scanline++
	v.mem.Write(LY_REGISTER, uint8(v.scanline))

	// check for LY=0 as well
	v.checkInterruption(true)
	v.lastComparison = false
}

func (v *Video) scan(c *Cpu) {

	// was disabled, now we reset
	if v.disabled && v.mem.Read(LCDC_REGISTER)&0x80 > 0 {
		lcds := v.mem.Read(LCD_REGISTER)
		log.Printf("REENABLING PPU %.8b PC=0x%X\n", lcds, c.pc)
		v.tick = 0
		v.total = 0
		v.total2 = 0
		v.delay = 0
		v.setMode(2)
		v.buffer = make([]Sprite, 0)
		v.currentOamAddr = 0
		v.disabled = false
		v.checkInterruption(true)
		return
	}

	// LCD/PPU disabled
	if v.mem.Read(LCDC_REGISTER)&0x80 == 0x0 {
		if !v.disabled {
			log.Printf("DISABLING PPU\n")
			v.scanline = 0
			v.scancolumn = 0
			v.setMode(0)
			v.mem.Write(LY_REGISTER, uint8(v.scanline))
			v.disabled = true
		}
		return
	}

	v.checkInterruption(false)

	v.total++
	v.total2++

	if c.debug {
		log.Printf("LCDC=%.8b STAT=%.8b, LY=%d, LYC=%d\n", v.mem.Read(LCDC_REGISTER), v.mem.Read(LCD_REGISTER), v.scanline, v.mem.Read(LYC_REGISTER))
	}

	// used to delay PPU by n ticks
	if v.delay > 0 {
		v.delay--
		if v.delay == 0 {
			if v.nextMode != v.mode {
				v.setMode(v.nextMode)
			}
		} else {
			return
		}
	}

	// process h-blank
	if v.mode == 0 {
		v.delay = 51
		if v.scanline < 144 {
			v.scancolumn = 0
			v.nextMode = 2
		} else {
			v.nextMode = 1
		}
	}

	// process v-blank
	if v.mode == 1 {
		if v.scanline == 153 {
			v.total2 = 0
			v.scanline = 0
			v.mem.Write(LY_REGISTER, uint8(v.scanline))
			v.scancolumn = 0
			v.delay = 114
			v.nextMode = 2
		} else {
			// advance LY
			v.advanceLy(c)
			v.delay = 114
			v.nextMode = 1
		}
	}

	// MODE 2 handling
	if v.mode == 2 {

		v.tick++

		// two sprites every m-cycle (20 m-cycles to finish mode 2)
		for range 2 {
			// reading sprites from OAM
			sprite := v.readOAMSprite(OAM_MEMORY_START)
			if v.shouldAddToBuffer(sprite) {
				v.buffer = append(v.buffer, sprite)
			}
		}

		// 20 m-cycles
		if v.tick == 21 {

			// reset oam address counter
			v.currentOamAddr = 0

			sort.SliceStable(v.buffer, func(i, j int) bool {
				return v.buffer[i].xPos > v.buffer[j].xPos
			})

			// drawing
			v.setMode(3)

			v.tick = 0
		}
	}

	// Mode 3 DRAWING
	if v.mode == 3 {

		// 4 dots per m-cycle
		for range 4 {
			bgPx, obgPx := v.fetchBackground()
			bgWd, obgWd, hasWindow := v.fetchWindow()

			// default is to display BG or Window, (can be overriden by sprites below)
			if hasWindow {
				v.videoMemory[v.scanline][v.scancolumn] = bgWd
			} else {
				v.videoMemory[v.scanline][v.scancolumn] = bgPx
			}

			for _, o := range v.buffer {

				if v.scancolumn >= int(o.xPos) && int(o.xPos+8) > v.scancolumn {

					// obj is disabled
					lcdc := v.mem.Read(LCDC_REGISTER)
					if lcdc&0x2 == 0 {
						continue
					}

					sprite := v.fetchTile(o.tile, 1, v.height())

					// vertical flip
					if o.flags&0x40 > 0 {
						cmp := sprite
						for fy := range v.height() {
							sprite[(v.height()-1)-fy] = cmp[fy]
						}
					}

					// horizontal flip
					if o.flags&0x20 > 0 {
						cmp := sprite
						for fy := range v.height() {
							for fx := range 8 {
								sprite[fy][7-fx] = cmp[fy][fx]
							}
						}
					}

					py := (v.scanline - int(o.yPos)) % int(v.height())
					px := (v.scancolumn - int(o.xPos)) % 8

					objPx := sprite[py][px]

					if objPx != 0x0 {
						// OBP0
						addr := Word(0xFF48)

						// OBP1
						if o.flags&0x10 > 0 {
							addr = 0xFF49
						}

						pixel := sprite[py][px]

						// BG-OVER-OBJ priority
						bgOverObj := o.flags&0x80 > 0

						// obj color palette
						palette := v.mem.Read(addr)

						// apply palette
						color := Pixel(palette & ((0x3) << (pixel * 2)) >> (pixel * 2))

						// bg over obj AND bg original color > 0
						if !bgOverObj || ((!hasWindow && obgPx == 0) || (hasWindow && obgWd == 0)) {
							v.videoMemory[v.scanline][v.scancolumn] = Pixel(color)
						}

						// stop process
						break
					}
				}
			}

			v.scancolumn++
			if v.scancolumn == 160 {
				// reset buffer
				v.buffer = make([]Sprite, 0)
				// advance LY
				v.advanceLy(c)
				v.delay = 4
				v.nextMode = 0
				v.scancolumn = 0
			}
		}
	}
}
