package ppu

import (
	"gonesem/internal/cartridge"
	"image"
	"image/color"
)

type PPU struct {
	ctrl   Ctrl   // Control register
	mask   Mask   // Mask register
	status Status // Status register

	scanline int16 // Current display scanline
	cycle    int16 // Offest into scanline giving current pixel

	// 15-bit VRAM and temporary VRAM address memory layout
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

	backgroundPixel   uint8
	backgroundPalette uint8

	spriteScanlineY          [8]uint8
	spriteScanlinePattern    [8]uint8
	spriteScanlineAttributes [8]uint8
	spriteScanlineX          [8]uint8

	spriteCount uint8

	spritePatternShiftRegistersLow  [8]uint8
	spritePatternShiftRegistersHigh [8]uint8

	foregroundPixel    uint8
	foregroundPalette  uint8
	foregroundPriority uint8

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
				if ppu.isRenderingEnabled() {
					ppu.incrementScrollX()
				}
			}
		}

		if ppu.cycle == 256 {
			if ppu.isRenderingEnabled() {
				ppu.incrementScrollY()
			}
		}

		if ppu.cycle == 257 {
			ppu.loadShiftRegisters()

			if ppu.isRenderingEnabled() {
				ppu.copyAddressX()
			}
		}

		if ppu.cycle == 338 || ppu.cycle == 340 {
			ppu.getNametableByte()
		}

		if ppu.scanline == -1 && ppu.cycle >= 280 && ppu.cycle < 305 {
			if ppu.isRenderingEnabled() {
				ppu.copyAddressY()
			}
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

	ppu.getBackgroundPixelData()
	ppu.getForegroundPixelData()
	ppu.renderFinalPixel()

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

func (ppu *PPU) TransferDMAData(addr uint8, value uint8) {
	ppu.oamData[addr] = value
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
