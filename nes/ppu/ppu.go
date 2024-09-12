package ppu

import (
	"gonesem/nes/cartridge"
)

type PPU struct {
	ctrl   Ctrl   // Control register
	mask   Mask   // Mask register
	status Status // Status register

	scanline int16 // Current display scanline
	cycle    int16 // Offest into scanline giving current pixel

	memoryAddress uint16 // CPU -> PPU data read/write
	addressLatch  bool   // HI/LO byte PPU write address latch

	dataBuffer uint8 // Temporary databuffer used in 1-cycle PPU data read delay

	EmitNMI bool

	nameTable    [2048]uint8
	paletteTable [32]uint8
	cartridge    *cartridge.Cartridge
}

func NewPPU(cartridge *cartridge.Cartridge) *PPU {
	return &PPU{cartridge: cartridge, addressLatch: false}
}

/*
*
Used by the CPU to read information from the PPU's registers or memory
i.e. the connection from the CPU to PPU via the NES's main bus

NOTE. It is important to keep in mind that the act of the CPU reading
from the PPU can transform the state of the PPU, e.g. The CPU reading
the status register via address 0x2002 will cause the the vertical blank
flag in the status register and the address latch register to become unset.
*/
func (ppu *PPU) Read(addr uint16) uint8 {
	var value uint8 = 0

	switch addr {
	case 0x0000: // Control
		break
	case 0x0001: // Mask
		break
	case 0x0002: // Status
		value = (uint8(ppu.status) & 0xE0) | (ppu.dataBuffer & 0x1F)

		ppu.setStatus(StatusVerticalBlank, false)
		ppu.addressLatch = false
	case 0x0003: // OAM Address
		break
	case 0x0004: // OAM Data
		break
	case 0x0005: // Scroll
		break
	case 0x0006: // PPU Address
		break
	case 0x0007: // PPU Data
		value = ppu.dataBuffer
		ppu.dataBuffer = ppu.readMemory(ppu.memoryAddress)

		if ppu.memoryAddress > 0x3F00 {
			value = ppu.dataBuffer
		}

		ppu.memoryAddress++
	}

	return value
}

/*
*
Used by the CPU to write information to the PPU's registers or memory
i.e. the connection from the CPU to PPU via the NES's main bus
*/
func (ppu *PPU) Write(addr uint16, value uint8) {
	switch addr {
	case 0x0000: // Control
		ppu.ctrl = Ctrl(value)
	case 0x0001: // Mask
		ppu.mask = Mask(value)
	case 0x0002: // Status
		break
	case 0x0003: // OAM Address
		break
	case 0x0004: // OAM Data
		break
	case 0x0005: // Scroll
		break
	case 0x0006: // PPU Address
		if !ppu.addressLatch {
			ppu.memoryAddress = (ppu.memoryAddress & 0x00FF) | (uint16(value) << 8)
			ppu.addressLatch = true
		} else {
			ppu.memoryAddress = (ppu.memoryAddress & 0xFF00) | uint16(value)
			ppu.addressLatch = false
		}
	case 0x0007: // PPU Data
		ppu.writeMemory(ppu.memoryAddress, value)
		ppu.memoryAddress++
	}
}

/*
*
Used for reading from PPU's internal video memory, used in conjunction with
writeMemory method to represent the PPU's internal bus and the memory available on that.
*/
func (ppu *PPU) readMemory(addr uint16) uint8 {
	switch {
	// Pattern memory address space, i.e. CHR memory found on cartidge
	case addr <= 0x1FFF:
		return ppu.cartridge.CHRRead(addr)
	// Name table address space
	case addr >= 0x2000 && addr <= 0x3EFF:
		return ppu.nameTable[addr%2048]
	// Palette table address sapce
	case addr >= 0x3F00 && addr <= 0x3FFF:
		addr = (addr - 0x3F00) % 32

		if addr == 0x0010 || addr == 0x0014 || addr == 0x0018 || addr == 0x001C {
			addr -= 0x0010
		}

		return ppu.paletteTable[addr]
	}

	return 0
}

/*
*
Used for writing to PPU's internal video memory, used in conjunction with
readMemory method to represent the PPU's internal bus and the memory available on that.
*/
func (ppu *PPU) writeMemory(addr uint16, value uint8) {
	switch {
	// Pattern memory address space, i.e. CHR memory found on cartidge
	// NOTE. Generally the cartridge contains ROM, however writes can be done
	// in cases where the cartridge also contains CHR RAM.
	case addr <= 0x1FFF:
		ppu.cartridge.CHRWrite(addr, value)
	// Palette table address sapce
	case addr >= 0x2000 && addr <= 0x3EFF:
		ppu.nameTable[addr%2048] = value
	// Palette table address sapce
	case addr >= 0x3F00 && addr <= 0x3FFF:
		addr = (addr - 0x3F00) % 32

		if addr == 0x0010 || addr == 0x0014 || addr == 0x0018 || addr == 0x001C {
			addr -= 0x0010
		}

		ppu.paletteTable[addr] = value
	}
}

func (ppu *PPU) Clock() {
	// ------------------- //
	// Pre-render scanline //
	// ------------------- //

	if ppu.scanline == -1 && ppu.cycle == 1 {
		ppu.setStatus(StatusVerticalBlank, false)
	}

	// --------------- //
	// Render scanline //
	// --------------- //

	// TODO

	// --------------------- //
	// Post-render scanlines //
	// --------------------- //

	if ppu.scanline == 241 && ppu.cycle == 1 {
		ppu.setStatus(StatusVerticalBlank, true)

		if ppu.getCtrl(CtrlGenerateNMI) {
			ppu.EmitNMI = true
		}
	}
}
