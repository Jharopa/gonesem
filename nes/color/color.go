package color

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
)

func NewColorPalette(palFilePath string) ([64]color.RGBA, error) {
	var paletteColors [64]color.RGBA

	// Load color data
	palFile, err := os.Open(palFilePath)

	if err != nil {
		log.Printf("Failed to open palette file %s: %s", palFilePath, err)
		return paletteColors, err
	}

	stat, err := palFile.Stat()

	if err != nil {
		log.Printf("Failed to retrieve palette file stats for %s: %s", palFilePath, err)
		return paletteColors, err
	}

	if stat.Size() != 64 {
		log.Printf("Palette data from palette file %s is not 64 bytes in size", palFilePath)
		err := fmt.Errorf("palette file %d is %s bytes in size, should be 64", stat.Size(), palFilePath)
		return paletteColors, err
	}

	palleteFileColors := make([]byte, stat.Size())

	_, err = bufio.NewReader(palFile).Read(palleteFileColors)

	if err != nil && err != io.EOF {
		log.Printf("Failed to read %s palette file into colors buffer: %s", palFilePath, err)
		return paletteColors, err
	}

	// Populate palette colors
	for i := 0; i < len(palleteFileColors); i += 3 {
		paletteColors[i] = color.RGBA{
			R: palleteFileColors[i],
			G: palleteFileColors[i+1],
			B: palleteFileColors[i+2],
			A: 0xFF,
		}
	}

	return paletteColors, nil
}
