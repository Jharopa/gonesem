package ppu

import (
	"gonesem/nes/cartridge"
	"gonesem/nes/memory"
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

	nameTableByte        uint8
	attributeTableByte   uint8
	patternTableByteLow  uint8
	patternTableByteHigh uint8

	patternShiftResgisterLow    uint16
	patternShiftResgisterHigh   uint16
	attributeShiftResgisterLow  uint16
	attributeShiftResgisterHigh uint16

	EmitNMI bool

	nameTable    [2048]uint8
	paletteTable [32]uint8
	colorPalette [64]color.RGBA
	cartridge    *cartridge.Cartridge

	frame         *image.RGBA
	FrameComplete bool
}

func NewPPU(cartridge *cartridge.Cartridge, colorPalette [64]color.RGBA) *PPU {
	return &PPU{
		cartridge:     cartridge,
		colorPalette:  colorPalette,
		addressLatch:  false,
		frame:         image.NewRGBA(image.Rect(0, 0, 256, 240)),
		FrameComplete: false,
	}
}

/*
Used by the CPU to read information from the PPU's registers or memory
i.e. the connection from the CPU to PPU via the NES's main bus

NOTE. It is important to keep in mind that the act of the CPU reading
from the PPU can transform the state of the PPU, e.g. The CPU reading
the status register via address 0x2002 will cause the the vertical blank
flag in the status register and the address latch register to become unset.
*/
func (ppu *PPU) CPURead(addr uint16) uint8 {
	var value uint8 = 0

	switch addr {
	case 0x0000: // PPUCTRL $2000
		break
	case 0x0001: // PPUMASK $2001
		break
	case 0x0002: // PPUSTATUS $2002
		value = (uint8(ppu.status) & 0xE0) | (ppu.dataBuffer & 0x1F)

		ppu.setStatus(StatusVerticalBlank, false)
		ppu.addressLatch = false
	case 0x0003: // OAMADDR $2003
		break
	case 0x0004: // OAMDATA $2004
		break
	case 0x0005: // PPUSCROLL $2005
		break
	case 0x0006: // PPUADDR $2006
		break
	case 0x0007: // PPUDATA $2007
		value = ppu.dataBuffer
		ppu.dataBuffer = ppu.read(uint16(ppu.vramAddr))

		if ppu.vramAddr >= 0x3F00 {
			value = ppu.dataBuffer
		}

		ppu.incrementAddress()
	}

	return value
}

/*
Used by the CPU to write information to the PPU's registers or memory
i.e. the connection from the CPU to PPU via the NES's main bus
*/
func (ppu *PPU) CPUWrite(addr uint16, value uint8) {
	switch addr {
	case 0x0000: // PPUCTRL $2000
		ppu.ctrl = Ctrl(value)
		ppu.tramAddr = (ppu.tramAddr & 0xF3FF) | uint16(value&0x03)<<10
	case 0x0001: // PPUMASK $2001
		ppu.mask = Mask(value)
	case 0x0002: // PPUSTATUS $2002
		break
	case 0x0003: // OAMADDR $2003
		break
	case 0x0004: // OAMDATA $2004
		break
	case 0x0005: // PPUSCROLL $2005
		if !ppu.addressLatch {
			ppu.tramAddr = (ppu.tramAddr & 0xFFE0) | uint16(value)>>3
			ppu.fineX = value & 0x07
			ppu.addressLatch = true
		} else {
			ppu.tramAddr = (ppu.tramAddr & 0xFC1F) | uint16(value&0xF8)<<2
			ppu.tramAddr = (ppu.tramAddr & 0x8FFF) | uint16(value&0x07)<<12
			ppu.addressLatch = false
		}
	case 0x0006: // PPUADDR $2006
		if !ppu.addressLatch {
			ppu.tramAddr = (ppu.tramAddr & 0x00FF) | uint16(value&0x3F)<<8
			ppu.addressLatch = true
		} else {
			ppu.tramAddr = (ppu.tramAddr & 0xFF00) | uint16(value)
			ppu.vramAddr = ppu.tramAddr
			ppu.addressLatch = false
		}
	case 0x0007: // PPUDATA $2007
		ppu.write(uint16(ppu.vramAddr), value)
		ppu.incrementAddress()
	}
}

/*
Increments the PPU's memory address by 32 if the Ctrl register's increment mode bit is set;
otherwise it will increment the PPU's memory address by 1
*/
func (ppu *PPU) incrementAddress() {
	if ppu.getCtrl(CtrlIncrementMode) {
		ppu.vramAddr += 0x20
	} else {
		ppu.vramAddr += 0x01
	}
}

/*
Used for reading from PPU's internal video memory, used in conjunction with
writeMemory method to represent the PPU's internal bus and the memory available on that.
*/
func (ppu *PPU) read(addr uint16) uint8 {
	switch {
	// Pattern memory address space, i.e. CHR memory found on cartidge
	case addr <= 0x1FFF:
		return ppu.cartridge.CHRRead(addr)
	// Nametable address space
	case addr >= 0x2000 && addr <= 0x3EFF:
		mirrorMode := ppu.cartridge.MirrorMode()
		nameTableIdx := memory.MirroredNametableIdx(mirrorMode, addr)

		return ppu.nameTable[nameTableIdx*0x400+addr%0x400]
	// Palette table address sapce
	case addr >= 0x3F00 && addr <= 0x3FFF:
		// If n = 2^x then y % n ≡ y & n-1
		// Example: 32 - 0010_0000
		//          31 - 0001_1111
		//          52 - 0011_0100
		// 			52 % 32 = 20 or 0001_0100
		// 			0011_0100 & 0001_1111 = 0001_0100 or 20
		// This works because when moduloing someting against some 2^x can be done
		// by taking all the lower order bits of that number below that 2^x value,
		// i.e. if n >= 0 then 2^(x + n) / 2^x = 0
		addr &= 0x1F

		if addr == 0x0010 || addr == 0x0014 || addr == 0x0018 || addr == 0x001C {
			addr -= 0x0010
		}

		return ppu.paletteTable[addr]
	}

	return 0
}

/*
Used for writing to PPU's internal video memory, used in conjunction with
readMemory method to represent the PPU's internal bus and the memory available on that.
*/
func (ppu *PPU) write(addr uint16, value uint8) {
	switch {
	// Pattern memory address space, i.e. CHR memory found on cartidge
	// NOTE. Generally the cartridge contains ROM, however writes can be done
	// in cases where the cartridge also contains CHR RAM.
	case addr <= 0x1FFF:
		ppu.cartridge.CHRWrite(addr, value)
	// Name table address sapce
	case addr >= 0x2000 && addr <= 0x3EFF:
		mirrorMode := ppu.cartridge.MirrorMode()
		nameTableIdx := memory.MirroredNametableIdx(mirrorMode, addr)

		ppu.nameTable[nameTableIdx*0x400+addr%0x400] = value
	// Palette table address sapce
	case addr >= 0x3F00 && addr <= 0x3FFF:
		addr &= 0x1F

		if addr == 0x0010 || addr == 0x0014 || addr == 0x0018 || addr == 0x001C {
			addr -= 0x0010
		}

		ppu.paletteTable[addr] = value
	}
}

func (ppu *PPU) getColourFromPaletteMemory(palette uint8, pixel uint8) color.RGBA {
	return ppu.colorPalette[ppu.read(0x3F00+uint16(palette<<2)+uint16(pixel))&0x3F]
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

func (ppu *PPU) getNametableByte() {
	ppu.nameTableByte = ppu.read(0x2000 | (uint16(ppu.vramAddr) & 0x0FFF))
}

func (ppu *PPU) getAttributeTableByte() {
	ppu.attributeTableByte = ppu.read(0x23C0 | (ppu.vramAddr & 0x0C00) | ((ppu.vramAddr >> 0x04) & 0x38) | ((ppu.vramAddr >> 0x02) & 0x07))

	if (ppu.vramAddr>>5)&0x02 == 0x02 {
		ppu.attributeTableByte >>= 4
	}

	if ppu.vramAddr&0x02 == 0x02 {
		ppu.attributeTableByte >>= 2
	}

	ppu.attributeTableByte &= 0x03
}

func (ppu *PPU) getPatternTableByte(low bool) uint8 {
	var (
		tableIdx    uint16
		fineY       uint16
		planeOffset uint16
		address     uint16
	)

	if !ppu.getCtrl(CtrlBackgroundTableAddres) {
		tableIdx = 0
	} else {
		tableIdx = 1
	}

	fineY = (ppu.vramAddr >> 12) & 0x07

	if low {
		planeOffset = 0
	} else {
		planeOffset = 8
	}

	address = tableIdx*0x1000 + uint16(ppu.nameTableByte)*16 + fineY + planeOffset

	return ppu.read(address)
}

func (ppu *PPU) getPatternTableByteLow() {
	ppu.patternTableByteLow = ppu.getPatternTableByte(true)
}

func (ppu *PPU) getPatternTableByteHigh() {
	ppu.patternTableByteHigh = ppu.getPatternTableByte(false)
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
	if ppu.getMask(MaskShowBackground) || ppu.getMask(MaskShowSprites) {
		ppu.patternShiftResgisterLow <<= 1
		ppu.patternShiftResgisterHigh <<= 1

		ppu.attributeShiftResgisterLow <<= 1
		ppu.attributeShiftResgisterHigh <<= 1
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
		}

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
	}

	if ppu.scanline == 241 && ppu.cycle == 1 {
		ppu.setStatus(StatusVerticalBlank, true)

		if ppu.getCtrl(CtrlGenerateNMI) {
			ppu.EmitNMI = true
		}
	}

	var (
		backgroundPixel   uint8
		backgroundPalette uint8
	)

	if ppu.getMask(MaskShowBackground) {
		var (
			bitMux      uint16 = 0x8000 >> ppu.fineX
			pixelLow    uint16
			pixelHigh   uint16
			paletteLow  uint16
			paletteHigh uint16
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

		backgroundPixel = (uint8(pixelHigh) << 1) | uint8(pixelLow)

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

		backgroundPalette = (uint8(paletteHigh) << 1) | uint8(paletteLow)

		// fmt.Printf("Scanline: %d - Cycle: %d Pixel: %d Palette: %d\n", ppu.scanline, ppu.cycle, backgroundPixel, backgroundPalette)
	}

	ppu.frame.Set(
		int(ppu.cycle),
		int(ppu.scanline),
		ppu.getColourFromPaletteMemory(backgroundPalette, backgroundPixel),
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
					pixel := (tileLSB & 0x01) + ((tileMSB & 0x01) << 1)
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
