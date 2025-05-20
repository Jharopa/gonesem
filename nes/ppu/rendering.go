package ppu

func (ppu *PPU) loadShiftRegisters() {
	ppu.patternShiftResgisterLow = (ppu.patternShiftResgisterLow & 0xFF00) | uint16(ppu.patternTableByteLow)
	ppu.patternShiftResgisterHigh = (ppu.patternShiftResgisterHigh & 0xFF00) | uint16(ppu.patternTableByteHigh)

	attributeTableLowBit := uint16(ppu.attributeTableByte & 0b01)
	attributeTableHighBit := uint16(ppu.attributeTableByte & 0b10)

	ppu.attributeShiftResgisterLow = (ppu.attributeShiftResgisterLow & 0xFF00)

	if attributeTableLowBit&0x01 == 0x01 {
		ppu.attributeShiftResgisterLow |= 0xFF
	}

	ppu.attributeShiftResgisterHigh = (ppu.attributeShiftResgisterHigh & 0xFF00)

	if attributeTableHighBit&0x02 == 0x02 {
		ppu.attributeShiftResgisterHigh |= 0xFF
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

func (ppu *PPU) isRenderingEnabled() bool {
	return ppu.getMask(MaskShowBackground) || ppu.getMask(MaskShowSprites)
}

func (ppu *PPU) getBackgroundPixelData() {
	ppu.backgroundPixel = 0x00
	ppu.backgroundPalette = 0x00

	if ppu.getMask(MaskShowBackground) {
		var (
			bitMux      uint16 = 0x8000 >> ppu.fineX
			pixelLow    uint8
			pixelHigh   uint8
			paletteLow  uint8
			paletteHigh uint8
		)

		pixelLow = uint8((ppu.patternShiftResgisterLow & bitMux) >> (15 - ppu.fineX))
		pixelHigh = uint8((ppu.patternShiftResgisterHigh & bitMux) >> (15 - ppu.fineX))

		ppu.backgroundPixel = uint8((pixelHigh << 1) | pixelLow)

		paletteLow = uint8((ppu.attributeShiftResgisterLow & bitMux) >> (15 - ppu.fineX))
		paletteHigh = uint8((ppu.attributeShiftResgisterHigh & bitMux) >> (15 - ppu.fineX))

		ppu.backgroundPalette = uint8((paletteHigh << 1) | paletteLow)
	}
}

func (ppu *PPU) getForegroundPixelData() {
	ppu.foregroundPixel = 0x00
	ppu.foregroundPalette = 0x00
	ppu.foregroundPriority = 0x00

	if ppu.getMask(MaskShowSprites) {
		var (
			pixelLow  uint8
			pixelHigh uint8
		)

		ppu.isSpiteZeroRendering = false

		for i := range ppu.spriteCount {
			if ppu.spriteScanlineX[i] == 0 {
				pixelLow = (ppu.spritePatternShiftRegistersLow[i] & 0x80) >> 7

				pixelHigh = (ppu.spritePatternShiftRegistersHigh[i] & 0x80) >> 7

				ppu.foregroundPixel = (pixelHigh << 1) | pixelLow

				ppu.foregroundPalette = (ppu.spriteScanlineAttributes[i] & 0x03) + 0x04

				ppu.foregroundPriority = (ppu.spriteScanlineAttributes[i] & 0x20) >> 5

				if ppu.foregroundPixel != 0 {
					if i == 0 {
						ppu.isSpiteZeroRendering = true
					}

					break
				}
			}
		}
	}
}

func (ppu *PPU) renderFinalPixel() {
	var (
		pixel   uint8
		palette uint8
	)

	if ppu.backgroundPixel == 0 && ppu.foregroundPixel == 0 {
		// Neither the background nor the foreground have visible pixel
		// Draw the palette's transparent background colour
		pixel = 0x00
		palette = 0x00
	} else if ppu.backgroundPixel == 0 && ppu.foregroundPixel > 0 {
		// The background pixel is transparent while The foregound pixel is visible
		// Draw the foreground pixel and palette
		pixel = ppu.foregroundPixel
		palette = ppu.foregroundPalette
	} else if ppu.backgroundPixel > 0 && ppu.foregroundPixel == 0 {
		// The background pixel is visible  while the foregound pixel is transparent
		// Draw the background pixel and palette
		pixel = ppu.backgroundPixel
		palette = ppu.backgroundPalette
	} else if ppu.backgroundPixel > 0 && ppu.foregroundPixel > 0 {
		// Both the background and foreground pixels are visisble
		// Draw pixel based on priority
		if ppu.foregroundPriority == 0 {
			pixel = ppu.foregroundPixel
			palette = ppu.foregroundPalette
		} else {
			pixel = ppu.backgroundPixel
			palette = ppu.backgroundPalette
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
}
