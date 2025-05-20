package ppu

import (
	"gonesem/nes/memory"
	"image/color"
)

func (ppu *PPU) getColourFromPaletteMemory(palette uint8, pixel uint8) color.RGBA {
	return ppu.colorPalette[ppu.read(0x3F00+uint16(palette<<2)+uint16(pixel))&0x3F]
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
		value = ppu.oamData[ppu.oamAddr]
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
		ppu.oamAddr = value
	case 0x0004: // OAMDATA $2004
		ppu.oamData[ppu.oamAddr] = value
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
Used for reading from PPU's internal video memory, used in conjunction with
write method to represent the PPU's internal bus and the memory available on that.
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
