package nes

import (
	"gonesem/nes/cartridge"
	"gonesem/nes/cpu"
	"gonesem/nes/ppu"
)

type NES struct {
	cpu       *cpu.CPU
	ppu       *ppu.PPU
	cartridge *cartridge.Cartridge

	ram [2048]uint8

	TotalCycles uint64
}

func NewNES(cartridge *cartridge.Cartridge) *NES {
	nes := &NES{TotalCycles: 0}

	cpu := cpu.NewCPU(nes)
	ppu := ppu.NewPPU(cartridge)

	nes.cpu = cpu
	nes.ppu = ppu

	return nes
}

func (nes *NES) Read(addr uint16) uint8 {
	switch {
	case addr <= 0x1FFF:
		return nes.ram[addr%0x0800]
	case addr >= 0x2000 && addr <= 0x3FFF:
		return nes.ppu.Read(addr % 0x0008)
	default:
		return nes.cartridge.PRGRead(addr)
	}
}

func (nes *NES) Write(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		nes.ram[addr%0x0800] = value
	case addr >= 0x2000 && addr <= 0x3FFF:
		nes.ppu.Write(addr%0x0008, value)
	default:
		nes.cartridge.PRGWrite(addr, value)
	}
}

func (nes *NES) Clock() {
	nes.ppu.Clock()

	if nes.TotalCycles%3 == 0 {
		nes.cpu.Clock()
	}

	if nes.ppu.EmitNMI {
		nes.cpu.NMI()
		nes.ppu.EmitNMI = false
	}

	nes.TotalCycles++
}
