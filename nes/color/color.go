package color

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"os"
)

func NewColorPalette(palFilePath string) ([64]color.RGBA, error) {
	var paletteColors [64]color.RGBA

	// Load color data
	palFile, err := os.Open(palFilePath)

	if err != nil {
		return paletteColors, fmt.Errorf("failed to open palette file %s: %s", palFilePath, err)
	}

	stat, err := palFile.Stat()

	if err != nil {
		return paletteColors, fmt.Errorf("failed to retrieve palette file stats for %s: %s", palFilePath, err)
	}

	if stat.Size() != 192 {
		err := fmt.Errorf("file %d is %s bytes in size, should be 192", stat.Size(), palFilePath)
		return paletteColors, err
	}

	palleteFileColors := make([]byte, stat.Size())

	_, err = bufio.NewReader(palFile).Read(palleteFileColors)

	if err != nil && err != io.EOF {
		return paletteColors, fmt.Errorf("failed to read %s palette file into colors buffer: %s", palFilePath, err)
	}

	// Populate palette colors
	for i := 0; i < len(palleteFileColors); i += 3 {
		paletteColors[i/3] = color.RGBA{
			R: palleteFileColors[i],
			G: palleteFileColors[i+1],
			B: palleteFileColors[i+2],
			A: 0xFF,
		}
	}

	return paletteColors, nil
}
