package main

import (
	"fmt"

	"github.com/Dudssource/shiny-cart/emulator"
)

func main() {
	fmt.Println("GB Classic Emulator")
	c := emulator.Cpu{}
	c.Start()
}
