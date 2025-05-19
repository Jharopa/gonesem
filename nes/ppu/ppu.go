package ppu

import (
	"gonesem/nes/cartridge"
	"image"
	"image/color"
)

type PPU struct {
	ctrl   Ctrl   // Control register
	mask   Mask   // Mask register
	status Status // Status register

	scanline int16 // Current display scanline
	cycle    int16 // Offest into scanline giving current pixel

	// VRAM and temporary VRAM memory layout
	// yyy NN YYYYY XXXXX
	// ||| || ||||| +++++-- coarse X scroll
	// ||| || +++++-------- coarse Y scroll
	// ||| |+-------------- nametable x select
	// ||| +--------------- nametable y select
	// +++----------------- fine Y scroll

	// Internal address registers
	vramAddr     uint16
	tramAddr     uint16
	fineX        uint8
	addressLatch bool // HI/LO byte PPU write address latch

	dataBuffer uint8 // Temporary databuffer used in 1 CPU cycle PPU data read delay

	oamAddr uint8

	nameTableByte        uint8
	attributeTableByte   uint8
	patternTableByteLow  uint8
	patternTableByteHigh uint8

	patternShiftResgisterLow    uint16
	patternShiftResgisterHigh   uint16
	attributeShiftResgisterLow  uint16
	attributeShiftResgisterHigh uint16

	spriteScanlineY          [8]uint8
	spriteScanlinePattern    [8]uint8
	spriteScanlineAttributes [8]uint8
	spriteScanlineX          [8]uint8

	spriteCount uint8

	spritePatternShiftRegistersLow  [8]uint8
	spritePatternShiftRegistersHigh [8]uint8

	canSpriteZeroHit     bool
	isSpiteZeroRendering bool

	EmitNMI bool

	nameTable    [2048]uint8
	paletteTable [32]uint8
	oamData      [256]uint8
	colorPalette [64]color.RGBA

	cartridge *cartridge.Cartridge

	frame         *image.RGBA
	FrameComplete bool
}

func NewPPU(cartridge *cartridge.Cartridge, colorPalette [64]color.RGBA) *PPU {
	return &PPU{
		cartridge:            cartridge,
		colorPalette:         colorPalette,
		addressLatch:         false,
		canSpriteZeroHit:     false,
		isSpiteZeroRendering: false,
		frame:                image.NewRGBA(image.Rect(0, 0, 256, 240)),
		FrameComplete:        false,
	}
}

/*
Increments current vram address along the scanline. If background or foreground
rendering are enabled, checks coarse x value in address is equal to 31 resetting
coarse x to 0 and flipping the x nametablebit;
Otherwise increments coarse x by 1.
*/
func (ppu *PPU) incrementScrollX() {
	if ppu.getMask(MaskShowBackground) || ppu.getMask(MaskShowSprites) {
		if ppu.vramAddr&0x1F == 0x1F {
			ppu.vramAddr &= 0xFFE0
			ppu.vramAddr ^= 0x0400
		} else {
			ppu.vramAddr++
		}
	}
}

/*
Increments current vram address down the scanline. If background or foreground
rendering are enabled, increments fine y by 1 if it is less than 7;
Otherwise accounting for the nametables attribute memory boundary,
either resets coarse y and/or flips the y nametable bit, or increments coarse y by 1.
*/
func (ppu *PPU) incrementScrollY() {
	if ppu.getMask(MaskShowBackground) || ppu.getMask(MaskShowSprites) {
		if ppu.vramAddr&0x7000 != 0x7000 {
			ppu.vramAddr += 0x1000
		} else {
			ppu.vramAddr &= 0x8FFF

			y := (ppu.vramAddr & 0x03E0) >> 5

			if y == 0x1D { // Coarse y is 29
				y = 0
				ppu.vramAddr ^= 0x0800
			} else if y == 0x1F { // Coarse y is 31
				y = 0
			} else {
				y++
			}

			ppu.vramAddr = (ppu.vramAddr & 0xFC1F) | (y << 5)
		}
	}
}

func (ppu *PPU) copyAddressX() {
	if ppu.getMask(MaskShowBackground) || ppu.getMask(MaskShowSprites) {
		ppu.vramAddr = (ppu.vramAddr & 0x7BE0) | (ppu.tramAddr & 0x041F)
	}
}

func (ppu *PPU) copyAddressY() {
	if ppu.getMask(MaskShowBackground) || ppu.getMask(MaskShowSprites) {
		ppu.vramAddr = (ppu.vramAddr & 0x041F) | (ppu.tramAddr & 0x7BE0)
	}
}

func (ppu *PPU) loadShiftRegisters() {
	ppu.patternShiftResgisterLow = (ppu.patternShiftResgisterLow & 0xFF00) | uint16(ppu.patternTableByteLow)
	ppu.patternShiftResgisterHigh = (ppu.patternShiftResgisterHigh & 0xFF00) | uint16(ppu.patternTableByteHigh)

	attributeTableLowBit := uint16(ppu.attributeTableByte & 0b01)
	attributeTableHighBit := uint16(ppu.attributeTableByte & 0b10)

	if attributeTableLowBit&0x01 == 0x01 {
		ppu.attributeShiftResgisterLow = (ppu.attributeShiftResgisterLow & 0xFF00) | 0xFF
	} else {
		ppu.attributeShiftResgisterLow = (ppu.attributeShiftResgisterLow & 0xFF00)
	}

	if attributeTableHighBit&0x02 == 0x02 {
		ppu.attributeShiftResgisterHigh = (ppu.attributeShiftResgisterHigh & 0xFF00) | 0xFF
	} else {
		ppu.attributeShiftResgisterHigh = (ppu.attributeShiftResgisterHigh & 0xFF00)
	}
}

func (ppu *PPU) updateShiftRegisters() {
	if ppu.getMask(MaskShowBackground) {
		ppu.patternShiftResgisterLow <<= 1
		ppu.patternShiftResgisterHigh <<= 1

		ppu.attributeShiftResgisterLow <<= 1
		ppu.attributeShiftResgisterHigh <<= 1
	}

	if ppu.getMask(MaskShowSprites) && ppu.cycle >= 1 && ppu.cycle < 258 {
		for i := range ppu.spriteCount {
			if ppu.spriteScanlineX[i] > 0 {
				ppu.spriteScanlineX[i]--
			} else {
				ppu.spritePatternShiftRegistersLow[i] <<= 1
				ppu.spritePatternShiftRegistersHigh[i] <<= 1
			}
		}
	}
}

func (ppu *PPU) Clock() {

	// ---------------- //
	// Noise generation //
	// ---------------- //

	// if (ppu.scanline >= 0 && ppu.scanline < 240) && (ppu.cycle >= 1 || ppu.cycle <= 256) {
	//     ppu.frame.Set(int(ppu.cycle), int(ppu.scanline), ppu.colorPalette[rand.Intn(len(ppu.colorPalette))])
	// }

	if ppu.scanline >= -1 && ppu.scanline < 240 {
		// Odd frame cycle skip
		if ppu.scanline == 0 && ppu.cycle == 0 {
			ppu.cycle = 1
		}

		if ppu.scanline == -1 && ppu.cycle == 1 {
			ppu.setStatus(StatusVerticalBlank, false)
			ppu.setStatus(StatusSpriteOverflow, false)
			ppu.setStatus(StatusSpriteZeroHit, false)

			for i := range 8 {
				ppu.spritePatternShiftRegistersLow[i] = 0
				ppu.spritePatternShiftRegistersHigh[i] = 0

				for i := range 8 {
					ppu.spriteScanlineY[i] = 0xFF
					ppu.spriteScanlinePattern[i] = 0xFF
					ppu.spriteScanlineAttributes[i] = 0xFF
					ppu.spriteScanlineX[i] = 0xFF
				}

			}
		}

		// -------------------- //
		// Background Rendering //
		// -------------------- //
		if (ppu.cycle >= 2 && ppu.cycle < 258) || (ppu.cycle >= 321 && ppu.cycle < 338) {
			ppu.updateShiftRegisters()

			switch (ppu.cycle - 1) % 8 {
			case 0:
				ppu.loadShiftRegisters()
				ppu.getNametableByte()
			case 2:
				ppu.getAttributeTableByte()
			case 4:
				ppu.getPatternTableByteLow()
			case 6:
				ppu.getPatternTableByteHigh()
			case 7:
				ppu.incrementScrollX()
			}
		}

		if ppu.cycle == 256 {
			ppu.incrementScrollY()
		}

		if ppu.cycle == 257 {
			ppu.loadShiftRegisters()
			ppu.copyAddressX()
		}

		if ppu.cycle == 338 || ppu.cycle == 340 {
			ppu.getNametableByte()
		}

		if ppu.scanline == -1 && ppu.cycle >= 280 && ppu.cycle < 305 {
			ppu.copyAddressY()
		}

		// -------------------- //
		// Foreground Rendering //
		// -------------------- //
		if ppu.cycle == 257 && ppu.scanline >= 0 {
			ppu.evaluateSprites()
		}

		if ppu.cycle == 340 {
			ppu.getSpritePatterns()
		}
	}

	if ppu.scanline == 241 && ppu.cycle == 1 {
		ppu.setStatus(StatusVerticalBlank, true)

		if ppu.getCtrl(CtrlGenerateNMI) {
			ppu.EmitNMI = true
		}
	}

	// -------------------- //
	// Background Rendering //
	// -------------------- //
	var (
		backgroundPixel   uint8
		backgroundPalette uint8
	)

	if ppu.getMask(MaskShowBackground) {
		var (
			bitMux      uint16 = 0x8000 >> ppu.fineX
			pixelLow    uint8
			pixelHigh   uint8
			paletteLow  uint8
			paletteHigh uint8
		)

		if (ppu.patternShiftResgisterLow & bitMux) > 0 {
			pixelLow = 1
		} else {
			pixelLow = 0
		}

		if (ppu.patternShiftResgisterHigh & bitMux) > 0 {
			pixelHigh = 1
		} else {
			pixelHigh = 0
		}

		backgroundPixel = uint8((pixelHigh << 1) | pixelLow)

		if (ppu.attributeShiftResgisterLow & bitMux) > 0 {
			paletteLow = 1
		} else {
			paletteLow = 0
		}

		if (ppu.attributeShiftResgisterHigh & bitMux) > 0 {
			paletteHigh = 1
		} else {
			paletteHigh = 0
		}

		backgroundPalette = uint8((paletteHigh << 1) | paletteLow)
	}

	// -------------------- //
	// Foreground Rendering //
	// -------------------- //
	var (
		foregroundPixel    uint8
		foregroundPalette  uint8
		foregroundPriority uint8
	)

	if ppu.getMask(MaskShowSprites) {
		var (
			pixelLow  uint8
			pixelHigh uint8
		)

		ppu.isSpiteZeroRendering = false

		for i := range ppu.spriteCount {
			if ppu.spriteScanlineX[i] == 0 {
				if (ppu.spritePatternShiftRegistersLow[i] & 0x80) > 0 {
					pixelLow = 1
				} else {
					pixelLow = 0
				}

				if (ppu.spritePatternShiftRegistersHigh[i] & 0x80) > 0 {
					pixelHigh = 1
				} else {
					pixelHigh = 0
				}

				foregroundPixel = (pixelHigh << 1) | pixelLow

				foregroundPalette = (ppu.spriteScanlineAttributes[i] & 0x03) + 0x04

				if (ppu.spriteScanlineAttributes[i] & 0x20) == 0 {
					foregroundPriority = 1
				} else {
					foregroundPriority = 0
				}

				if foregroundPixel != 0 {
					if i == 0 {
						ppu.isSpiteZeroRendering = true
					}

					break
				}
			}
		}
	}

	var (
		pixel   uint8
		palette uint8
	)

	if backgroundPixel == 0 && foregroundPixel == 0 {
		// Neither the background nor the foreground have visible pixel
		// Draw the palette's transparent background colour
		pixel = 0x00
		palette = 0x00
	} else if backgroundPixel == 0 && foregroundPixel > 0 {
		// The background pixel is transparent while The foregound pixel is visible
		// Draw the foreground pixel and palette
		pixel = foregroundPixel
		palette = foregroundPalette
	} else if backgroundPixel > 0 && foregroundPixel == 0 {
		// The background pixel is visible  while the foregound pixel is transparent
		// Draw the background pixel and palette
		pixel = backgroundPixel
		palette = backgroundPalette
	} else if backgroundPixel > 0 && foregroundPixel > 0 {
		// Both the background and foreground pixels are visisble
		// Draw pixel based on priority
		if foregroundPriority == 1 {
			pixel = foregroundPixel
			palette = foregroundPalette
		} else {
			pixel = backgroundPixel
			palette = backgroundPalette
		}

		if ppu.canSpriteZeroHit && ppu.isSpiteZeroRendering {
			if ppu.getMask(MaskShowBackground) && ppu.getMask(MaskShowSprites) {
				if !ppu.getMask(MaskShowBackgroundLeft) || ppu.getMask(MaskShowSpritesLeft) {
					if ppu.cycle >= 9 && ppu.cycle < 258 {
						ppu.setStatus(StatusSpriteZeroHit, true)
					}
				} else {
					if ppu.cycle >= 1 && ppu.cycle < 258 {
						ppu.setStatus(StatusSpriteZeroHit, true)
					}
				}
			}
		}
	}

	ppu.frame.Set(
		int(ppu.cycle-1),
		int(ppu.scanline),
		ppu.getColourFromPaletteMemory(palette, pixel),
	)

	ppu.cycle++

	if ppu.cycle >= 341 {
		ppu.cycle = 0

		ppu.scanline++

		if ppu.scanline >= 261 {
			ppu.scanline = -1
			ppu.FrameComplete = true
		}
	}
}

func (ppu *PPU) GetFrame() *image.RGBA {
	return ppu.frame
}

func (ppu *PPU) GetPatternTable(tableIndex uint8, paletteIndex uint8) *image.RGBA {
	var x, y, row, col uint16
	patternTableImage := image.NewRGBA(image.Rect(0, 0, 128, 128))

	for y = 0; y < 16; y++ {
		for x = 0; x < 16; x++ {
			tileOffset := y*256 + x*16

			for row = 0; row < 8; row++ {
				tileLSB := ppu.read(uint16(tableIndex)*0x1000 + tileOffset + row)
				tileMSB := ppu.read(uint16(tableIndex)*0x1000 + tileOffset + row + 8)

				for col = 0; col < 8; col++ {
					pixel := (tileLSB & 0x01 << 1) | ((tileMSB & 0x01) << 1)
					tileLSB >>= 1
					tileMSB >>= 1

					patternTableImage.Set(
						int(x*8+(7-col)),
						int(y*8+row),
						ppu.colorPalette[ppu.read(0x3F00+uint16(paletteIndex<<2)+uint16(pixel))&0x3F],
					)
				}
			}
		}
	}

	return patternTableImage
}
