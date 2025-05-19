package ppu

import (
	"math/bits"
)

/*
Finds up to the first eight sprites that intersect the next scanline,
copying that sprites OAM memory into the internal spriteScanline array.

Any sprites over the first eight are ignored, however if there are more than
eight sprites that intersect the next scanline, the sprite overflow bit in
the PPU's status regsiter is set.
*/
func (ppu *PPU) evaluateSprites() {
	for i := range 8 {
		ppu.spriteScanlineY[i] = 0xFF
		ppu.spriteScanlinePattern[i] = 0xFF
		ppu.spriteScanlineAttributes[i] = 0xFF
		ppu.spriteScanlineX[i] = 0xFF
	}

	ppu.spriteCount = 0

	for i := range 8 {
		ppu.spritePatternShiftRegistersLow[i] = 0
		ppu.spritePatternShiftRegistersHigh[i] = 0
	}

	ppu.canSpriteZeroHit = true

	for i := range 64 {
		if ppu.spriteCount > 8 {
			break
		}

		// Current scanline minus the current sprites y position
		row := ppu.scanline - int16(ppu.oamData[i*4])

		if row >= 0 && row < int16(ppu.getSpriteHeight()) {
			if ppu.spriteCount < 8 {
				if i == 0 {
					ppu.canSpriteZeroHit = true
				}

				ppu.spriteScanlineY[ppu.spriteCount] = ppu.oamData[i*4]            // Y position
				ppu.spriteScanlinePattern[ppu.spriteCount] = ppu.oamData[i*4+1]    // Index number
				ppu.spriteScanlineAttributes[ppu.spriteCount] = ppu.oamData[i*4+2] // Attributes
				ppu.spriteScanlineX[ppu.spriteCount] = ppu.oamData[i*4+3]          // X position

				ppu.spriteCount++
			}
		}
	}

	if ppu.spriteCount > 8 {
		ppu.setStatus(StatusSpriteOverflow, true)
	}
}

func (ppu *PPU) getSpritePatterns() {
	for i := range ppu.spriteCount {
		var (
			spritePatternBitsLow, spritePatternBitsHigh uint8
			spritePatternAddrLow, spritePatternAddrHigh uint16
		)

		// 8x8 Sprite Mode
		if ppu.getSpriteHeight() == 0x08 {
			patternTableAddr := ppu.getSpritePatternTableAddress()
			patternTableIdx := uint16(ppu.spriteScanlinePattern[i]) << 4
			row := uint16(int16(ppu.scanline) - int16(ppu.spriteScanlineY[i]))

			if (ppu.spriteScanlineAttributes[i] & 0x80) != 0x80 { // Sprite is not flipped veritcally
				spritePatternAddrLow = patternTableAddr | patternTableIdx | row
			} else { // Sprite is flipped veritcally
				spritePatternAddrLow = patternTableAddr | patternTableIdx | (7 - row)
			}
		} else { // 8x16 Sprite Mode
			patternTableAddr := uint16(ppu.spriteScanlinePattern[i]&0x01) << 12
			patternTableIdx := uint16(ppu.spriteScanlinePattern[i]) & 0xFE
			row := uint16(int16(ppu.scanline)-int16(ppu.spriteScanlineY[i])) & 0x07

			if (ppu.spriteScanlineAttributes[i] & 0x80) != 0x80 { // Sprite is not flipped veritcally
				if ppu.scanline-int16(ppu.spriteScanlineY[i]) < 8 {
					spritePatternAddrLow = patternTableAddr | (patternTableIdx << 4) | row
				} else {
					spritePatternAddrLow = patternTableAddr | ((patternTableIdx + 1) << 4) | row
				}
			} else { // Sprite is flipped veritcally
				if ppu.scanline-int16(ppu.spriteScanlineY[i]) < 8 {
					spritePatternAddrLow = patternTableAddr | (patternTableIdx << 4) | (7 - row)
				} else {
					spritePatternAddrLow = patternTableAddr | (patternTableIdx+1)<<4 | (7 - row)
				}
			}
		}

		spritePatternAddrHigh = spritePatternAddrLow + 0x08
		spritePatternBitsLow = ppu.read(spritePatternAddrLow)
		spritePatternBitsHigh = ppu.read(spritePatternAddrHigh)

		// Sprite is horizontal flipped
		if ppu.spriteScanlineAttributes[i]&0x40 > 0 {
			spritePatternBitsLow = bits.Reverse8(spritePatternBitsLow)
			spritePatternBitsHigh = bits.Reverse8(spritePatternBitsHigh)
		}

		ppu.spritePatternShiftRegistersLow[i] = spritePatternBitsLow
		ppu.spritePatternShiftRegistersHigh[i] = spritePatternBitsHigh
	}
}

func (ppu *PPU) getSpriteHeight() uint8 {
	if !ppu.getCtrl(CtrlSpriteSize) {
		return 0x08
	} else {
		return 0x10
	}
}

func (ppu *PPU) getSpritePatternTableAddress() uint16 {
	if !ppu.getCtrl(CtrlSpriteTableAddress) {
		return 0x00
	} else {
		return 0x1000
	}
}
