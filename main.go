package main

import (
	"gonesem/nes"
	"gonesem/nes/cartridge"
	"gonesem/nes/color"
	"log"
	"os"
)

func main() {
	nes, err := nesInit()

	if err != nil {
		log.Fatalf("Failed to initialize NES console: %s\n", err)

		os.Exit(1)
	}

	for {
		nes.Clock()
	}
}

func nesInit() (*nes.NES, error) {
	cartridge, err := cartridge.NewCartridge("./test/data/roms/Donkey Kong.nes")

	if err != nil {
		return nil, err
	}

	colorPalette, err := color.NewColorPalette("./test/data/pals/NESdev.pal")

	if err != nil {
		return nil, err
	}

	nes := nes.NewNES(cartridge, colorPalette)

	return nes, nil
}
