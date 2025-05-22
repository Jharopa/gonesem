package application

import (
	"gonesem/internal/cartridge"
	"gonesem/internal/color"
	"gonesem/internal/nes"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Application struct {
	nes          *nes.NES
	screenWidth  int
	screenHeight int
	scale        int
}

func NewApplication(options *Options) (*Application, error) {
	nes, err := initNes(options.rom, options.palette)

	if err != nil {
		return nil, err
	}

	rl.InitWindow(
		int32(256*options.scale),
		int32(240*options.scale),
		"GoNESEm",
	)
	rl.SetTraceLogLevel(rl.LogError)
	rl.SetTargetFPS(60)

	return &Application{
		nes:          nes,
		screenWidth:  256,
		screenHeight: 240,
		scale:        options.scale,
	}, nil
}

func initNes(romPath, palPath string) (*nes.NES, error) {
	cartridge, err := cartridge.NewCartridge(romPath)

	if err != nil {
		return nil, err
	}

	colorPalette, err := color.NewColorPalette(palPath)

	if err != nil {
		return nil, err
	}

	nes := nes.NewNES(cartridge, colorPalette)

	return nes, nil
}

func (app *Application) Run() {
	for !rl.WindowShouldClose() {
		for !app.nes.FrameComplete() {
			app.nes.Clock()
		}

		nesImage := rl.NewImageFromImage(app.nes.GetFrame())
		nesTexture := rl.LoadTextureFromImage(nesImage)

		rl.UnloadImage(nesImage)

		app.nes.ResetFrameComplete()

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		rl.DrawTexturePro(
			nesTexture,
			rl.NewRectangle(0, 0, float32(nesTexture.Width), float32(nesTexture.Height)),
			rl.NewRectangle(0, 0, float32(app.screenWidth*app.scale), float32(app.screenHeight*app.scale)),
			rl.NewVector2(0, 0),
			float32(0),
			rl.White,
		)

		rl.EndDrawing()

		rl.UnloadTexture(nesTexture)

		app.processInput()
	}
}

func (app *Application) processInput() {
	app.nes.Controller[0] = 0x00

	if rl.IsKeyDown(rl.KeySpace) { // A
		app.nes.Controller[0] |= 0x80
	} else {
		app.nes.Controller[0] |= 0x00
	}

	if rl.IsKeyDown(rl.KeyE) { // B
		app.nes.Controller[0] |= 0x40
	} else {
		app.nes.Controller[0] |= 0x00
	}

	if rl.IsKeyDown(rl.KeyLeftShift) { // Select
		app.nes.Controller[0] |= 0x20
	} else {
		app.nes.Controller[0] |= 0x00
	}

	if rl.IsKeyDown(rl.KeyEnter) { // Start
		app.nes.Controller[0] |= 0x10
	} else {
		app.nes.Controller[0] |= 0x00
	}

	if rl.IsKeyDown(rl.KeyW) { // Up
		app.nes.Controller[0] |= 0x08
	} else {
		app.nes.Controller[0] |= 0x00
	}

	if rl.IsKeyDown(rl.KeyS) { // Down
		app.nes.Controller[0] |= 0x04
	} else {
		app.nes.Controller[0] |= 0x00
	}

	if rl.IsKeyDown(rl.KeyA) { // Left
		app.nes.Controller[0] |= 0x02
	} else {
		app.nes.Controller[0] |= 0x00
	}

	if rl.IsKeyDown(rl.KeyD) { // Right
		app.nes.Controller[0] |= 0x01
	} else {
		app.nes.Controller[0] |= 0x00
	}
}
