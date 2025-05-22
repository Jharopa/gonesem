package application

import "flag"

type Options struct {
	rom     string
	palette string
	scale   int
}

func NewOptions() *Options {
	return &Options{}
}

func (app *Options) Parse() {
	flag.StringVar(&app.rom, "rom", "", "Path to the NES rom file")
	flag.StringVar(&app.palette, "palette", "", "Path the NES palette file")
	flag.IntVar(&app.scale, "scale", 3, "NES emulator display scale - Default: 3")

	flag.Parse()
}
