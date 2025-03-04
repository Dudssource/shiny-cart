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

type Video struct {
	w              int32
	h              int32
	internalWidth  int32
	internalHeight int32
	scaleFactor    int32

	mem            *Memory
	scanline       int
	scancolumn     int
	mode           uint8
	buffer         []Sprite
	tick           int
	delay          int
	nextMode       uint8
	disabled       bool
	lastComparison bool

	currentOamAddr Word

	videoMemory [144][160]Pixel

	ts     int
	t      time.Time
	total  int
	total2 int
}

func (v *Video) Draw(bitmap []byte, x, y int32, rowWiseBitmap bool) {

	//rl.BeginDrawing()
	//rl.ClearBackground(rl.RayWhite)

	//defer rl.EndDrawing()
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

// func (v *Video) fetchWindow() [256][256]Pixel {
// 	return v.fetchBackgroundOrWindowTiles(v.mem.Read(LCDC_REGISTER) & 0x40 >> 6)
// }

func (v *Video) fetchBackground() Pixel {
	return v.fetchBackgroundOrWindowTiles(v.mem.Read(LCDC_REGISTER) & 0x8 >> 3)
}

func (v *Video) fetchBackgroundOrWindowTiles(tileMap uint8) Pixel {
	mode := v.mem.Read(LCDC_REGISTER) & 0x10 >> 4
	address := 0x9800
	if tileMap == 1 {
		address = 0x9C00
	}

	y := v.scanline
	x := v.scancolumn

	scy := int(v.mem.Read(0xff42))
	scx := int(v.mem.Read(0xff43))

	tileY := (((scy / 8) + (y / 8)) % 32) & 31
	tileX := ((scx + (x / 8)) % 32) & 31
	tileAddress := Word(address + (tileY * 32) + tileX)
	tileNumber := v.mem.Read(tileAddress)
	tile := v.fetchTile(tileNumber, mode, 8)
	palette := v.mem.Read(0xFF47)
	pixel := tile[y%8][x%8]
	color := palette & ((0x3) << (pixel * 2)) >> (pixel * 2)
	return Pixel(color)
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
	rl.ClearBackground(rl.RayWhite)
	defer rl.EndDrawing()

	// https://www.deviantart.com/thewolfbunny64/art/Game-Boy-Palette-Lime-Midori-810574708
	// colors := map[Pixel]color.RGBA{
	// 	0: rl.NewColor(224, 235, 175, 255),
	// 	1: rl.NewColor(170, 207, 83, 255),
	// 	2: rl.NewColor(123, 141, 66, 255),
	// 	3: rl.NewColor(71, 89, 80, 255),
	// }

	// https://www.deviantart.com/thewolfbunny64/art/Game-Boy-Palette-Pokemon-Pinball-Ver-882658817
	colors := map[Pixel]color.RGBA{
		0: rl.NewColor(232, 248, 184, 255),
		1: rl.NewColor(160, 176, 80, 255),
		2: rl.NewColor(120, 96, 48, 255),
		3: rl.NewColor(24, 24, 32, 255),
	}

	// colors := map[Pixel]color.RGBA{
	// 	0: rl.RayWhite,
	// 	1: rl.LightGray,
	// 	2: rl.DarkGray,
	// 	3: rl.Black,
	// }

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
			//rl.DrawPixel(int32(x), int32(y), colors[pixel])
		}
	}
}

func (v *Video) checkInterruption(c *Cpu, resetCondition bool) {
	lcds := v.mem.Read(LCD_REGISTER)
	lyc := v.mem.Read(LYC_REGISTER)

	// LYC enabled, LYC=LY CONDITION IS MET, LY=LYC COMPARISON IS FALSE (IRQ BLOCKING)
	if lcds&0x40 > 0 && lyc == uint8(v.scanline) {
		if lcds&0x4 == 0 {
			iflag := v.mem.Read(INTERRUPT_FLAG) | 0x2
			v.mem.Write(INTERRUPT_FLAG, iflag)
			v.mem.Write(LCD_REGISTER, lcds|0x4)
			log.Printf("REQUESTING INTERRUPTION LYC=%d LY=%d LCDS=%.8b LCDC=%.8b IME=%d PC=0x%X\n", lyc, v.scanline, lcds, v.mem.Read(LCDC_REGISTER), c.ime, c.pc)
		}
		//v.lastComparison = true
	} else if resetCondition {
		// reset LY=LYC COMPARISON FLAG
		v.mem.Write(LCD_REGISTER, lcds&0xFB)
	}
}

func (v *Video) advanceLy(c *Cpu) {
	v.scanline++
	v.mem.Write(LY_REGISTER, uint8(v.scanline))
	// check for LY=0 as well
	v.checkInterruption(c, true)
	v.lastComparison = false
	if c.step {
		fmt.Printf("LY=%d\n", v.scanline)
	}
}

func (v *Video) scan(c *Cpu) {

	// was disabled, now we reset
	if v.disabled && v.mem.Read(LCDC_REGISTER)&0x80 > 0 {
		lcds := v.mem.Read(LCD_REGISTER)
		log.Printf("REENABLING PPU %.8b PC=0x%X >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n", lcds, c.pc)
		v.scanline = 0
		v.scancolumn = 0
		v.mode = 2
		v.tick = 0
		v.total = 0
		v.total2 = 0
		v.delay = 0
		v.buffer = make([]Sprite, 0)
		v.currentOamAddr = 0
		v.disabled = false
		v.checkInterruption(c, true)
		return
	}

	// LCD/PPU disabled
	if v.mem.Read(LCDC_REGISTER)&0x80 == 0x0 {
		if !v.disabled {
			log.Printf("DISABLING PPU\n")
			v.disabled = true
		}
		return
	}

	v.checkInterruption(c, false)

	v.total++
	v.total2++

	if c.step {
		fmt.Printf("LCDC=%.8b STAT=%.8b, LY=%d, LYC=%d\n", v.mem.Read(LCDC_REGISTER), v.mem.Read(LCD_REGISTER), v.scanline, v.mem.Read(LYC_REGISTER))
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
			//fmt.Printf("TOTAL2 %d\n", v.total2)
			v.total2 = 0
			v.scanline = 0
			v.scancolumn = 0
			v.setMode(2)
		} else {
			// advance LY
			v.advanceLy(c)
			if v.scanline > 144 {
				v.delay = 114
				v.nextMode = 1
			}
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
		if v.tick == 20 {

			// reset oam address counter
			v.currentOamAddr = 0

			sort.SliceStable(v.buffer, func(i, j int) bool {
				return v.buffer[i].xPos < v.buffer[j].xPos
			})

			// if len(v.buffer) > 0 {
			// 	fmt.Printf("LY=%d %v (%d)\n", v.scanline, v.buffer, len(v.buffer))
			// }

			// drawing
			v.setMode(3)

			v.tick = 0
		}
	}

	// Mode 3 DRAWING
	if v.mode == 3 {

		// addr := Word(OAM_MEMORY_START)
		// currentOamAddr := Word(0)
		// for k := range 40 {

		// 	baddr := addr + currentOamAddr
		// 	sprite := Sprite{}
		// 	sprite.yPos = v.mem.Read(Word(addr+currentOamAddr)) - 16
		// 	currentOamAddr++
		// 	sprite.xPos = v.mem.Read(Word(addr+v.currentOamAddr)) - 8
		// 	currentOamAddr++
		// 	sprite.tile = int(v.mem.Read(Word(addr + v.currentOamAddr)))
		// 	currentOamAddr++
		// 	sprite.flags = v.mem.Read(Word(addr + v.currentOamAddr))
		// 	currentOamAddr++

		// 	if v.scanline == 32 {
		// 		fmt.Printf("BUFFER o=%d addr=%X oy=%d ox=%d oflags=%.8b otile=%d y=%d x=%d h=%d\n", k, baddr, sprite.yPos, sprite.xPos, sprite.flags, sprite.tile, v.scanline, v.scancolumn, v.height())
		// 	}
		// }

		// 4 dots per m-cycle
		for range 4 {
			processed := false
			bgPx := v.fetchBackground()

			for _, o := range v.buffer {

				// if v.scancolumn > int(o.xPos+8) && int(o.xPos+15) > v.scancolumn {
				if v.scancolumn >= int(o.xPos) && int(o.xPos+(v.height())) > v.scancolumn {

					sprite := v.fetchTile(o.tile, 1, v.height())

					// vertical flip
					if o.flags&0x40 > 0 {
						cmp := sprite
						for fy := range 8 {
							sprite[7-fy] = cmp[fy]
						}
					}

					// horizontal flip
					if o.flags&0x20 > 0 {
						cmp := sprite
						for fy := range 8 {
							for fx := range 8 {
								sprite[fy][7-fx] = cmp[fy][fx]
							}
						}
					}

					py := (v.scanline - int(o.yPos)) % 8
					px := (v.scancolumn - int(o.xPos)) % 8
					pixel := sprite[py][px]

					color := pixel

					// custom alpha indicator
					if pixel == 0x0 {
						color = 100
					} else {
						// OBP0
						addr := Word(0xFF48)
						// OBP1
						if o.flags&0x10 > 0 {
							addr = 0xFF49
						}
						palette := v.mem.Read(addr)
						color = Pixel(palette & ((0x3) << (pixel * 2)) >> (pixel * 2))
					}

					if color == 100 || (bgPx > 0 && o.flags&0x80 > 0) {
						v.videoMemory[v.scanline][v.scancolumn] = bgPx
					} else {

						v.videoMemory[v.scanline][v.scancolumn] = Pixel(color)

						//fmt.Printf("SPRITE oy=%d ox=%d oflags=%.8b otile=%d y=%d x=%d h=%d py=%d px=%d c=%d \n", o.yPos, o.xPos, o.flags, o.tile, v.scanline, v.scancolumn, v.height(), py, px, color)
					}

					// found pixel during merge
					processed = true
					break
				}
			}

			if !processed {
				v.videoMemory[v.scanline][v.scancolumn] = bgPx
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
