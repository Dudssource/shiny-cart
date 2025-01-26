package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Dudssource/shiny-cart/emulator"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.LUTC)
	fmt.Println("GB Classic Emulator")

	flag.Args()
	file := flag.String("f", "", "ROM `file` location")
	debug := flag.Bool("d", false, "Debug mode")
	step := flag.Bool("s", false, "Step mode")
	interval := flag.Duration("t", 500*time.Millisecond, "Machine cycle interval")
	flag.Parse()

	// validate args
	if file == nil || *file == "" {
		flag.Usage()
		os.Exit(1)
		return
	}

	// emulator
	g := emulator.NewGameBoy(*debug, *step)

	// load ROM
	if err := g.Load(*file); err != nil {
		panic(err)
	}

	// game Loop
	g.Loop(*interval)
}
