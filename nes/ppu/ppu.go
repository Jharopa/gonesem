package ppu

import "gonesem/nes/cartridge"

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
	CR Ctrl
	MR Mask
	SR Status

	cartridge *cartridge.Cartridge
}

func NewPPU(cartridge *cartridge.Cartridge) *PPU {
	return &PPU{cartridge: cartridge}
}

func (ppu *PPU) Read(addr uint16) uint8 {
	switch addr {
	case 0x0000:
		break
	case 0x0001:
		break
	case 0x0002:
		break
	case 0x0003:
		break
	case 0x0004:
		break
	case 0x0005:
		break
	case 0x0006:
		break
	case 0x0007:
		break
	}

	return 0
}

func (ppu *PPU) Write(addr uint16, value uint8) {
	switch addr {
	case 0x0000:
		break
	case 0x0001:
		break
	case 0x0002:
		break
	case 0x0003:
		break
	case 0x0004:
		break
	case 0x0005:
		break
	case 0x0006:
		break
	case 0x0007:
		break
	}
}

func (ppu *PPU) Clock() {

}
