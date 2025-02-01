package emulator

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Video struct {
	width          int32
	height         int32
	internalWidth  int32
	internalHeight int32
	videoMemory    [][]uint8
	scaleFactor    int32
}

func (v *Video) Draw(bitmap []byte, x, y int32, rowWiseBitmap bool) {

	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	defer rl.EndDrawing()
}

func (v *Video) init(width, height int32) {
	v.internalWidth = 64
	v.internalHeight = 32
	v.width = width
	v.height = height
	v.scaleFactor = v.width / v.internalWidth

	v.videoMemory = make([][]uint8, v.internalHeight)
	for i := range v.videoMemory {
		v.videoMemory[i] = make([]uint8, v.internalWidth)
	}

	rl.InitWindow(v.width, v.height, "GameBoy-DMG Emulator")
	rl.SetTargetFPS(60)
}
