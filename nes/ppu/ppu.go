package ppu

import (
	"gonesem/nes/cartridge"
)

type (
	Ctrl   uint8
	Mask   uint8
	Status uint8
)

const (
	CtrlNametableAddressX     Ctrl = 1 << iota // N 1: Add 256 to X scroll position
	CtrlNametableAddressY                      // N 1: Add 240 to Y scroll position
	CtrlIncrementMode                          // I 0: Add 1 across; 1: Add 32 down
	CtrlSpriteTableAddress                     // S 0: Address $0000; 1: Address $1000; ignore in 8x16
	CtrlBackgroundTableAddres                  // B 0: Address $0000; 1: Address $1000
	CtrlSpriteSize                             // H 0: 8x8 pixles; 1: 8x16 pixels
	CtrlMasterSlaveMode                        // P 0: read backfrop from EXT pins; 1: output color on EXT pins
	CtrlGenerateNMI                            // V 0: off; 1: on
)

const (
	MaskGreyscale          Mask = 1 << iota // g
	MaskShowBackgroundLeft                  // m Show background in the leftmost 8 pixles of screen
	MaskShowSpritesLeft                     // M Show sprites in the leftmost 8 pixles of screen
	MaskShowBackground                      // b
	MaskShowSprites                         // s
	MaskEmphasizeRed                        // R
	MaskEmphasizeGreen                      // G
	MaskEmphasizeBlue                       // B
)

const (
	StatusOpenBus       Status = 1 << 5 // O
	StatusSpriteZeroHit Status = 1 << 6 // S
	StatusVerticalBlank Status = 1 << 7 // V
)

type PPU struct {
	ctrl   Ctrl   // Control register
	mask   Mask   // Mask register
	status Status // Status register

	memoryAddress uint16 // CPU -> PPU data read/write
	addressLatch  bool   // HI/LO byte PPU write address latch

	dataBuffer uint8 // Temporary databuffer used in 1-cycle PPU data read delay

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
flag in that register to become unset.
*/
func (ppu *PPU) Read(addr uint16) uint8 {
	switch addr {
	case 0x0000: // Control
		break
	case 0x0001: // Mask
		break
	case 0x0002: // Status
		value := uint8(ppu.status) | (ppu.dataBuffer & 0x1F)

		ppu.status &^= StatusVerticalBlank

		return value
	case 0x0003: // OAM Address
		break
	case 0x0004: // OAM Data
		break
	case 0x0005: // Scroll
		break
	case 0x0006: // PPU Address
		break
	case 0x0007: // PPU Data
		if ppu.memoryAddress > 0x3F00 {
			return ppu.dataBuffer
		}

		value := ppu.dataBuffer
		ppu.dataBuffer = ppu.readMemory(ppu.memoryAddress)

		return value
	}

	return 0
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
			ppu.memoryAddress = uint16(value)
			ppu.addressLatch = true
		} else {
			ppu.memoryAddress = (ppu.memoryAddress << 8) | uint16(value)
			ppu.addressLatch = false
		}
	case 0x0007: // PPU Data
		ppu.writeMemory(ppu.memoryAddress, value)
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

}
