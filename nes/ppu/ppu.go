package ppu

import "gonesem/nes/cartridge"

type PPU struct {
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
