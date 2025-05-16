package nes

import (
	"gonesem/nes/cartridge"
	"gonesem/nes/cpu"
	"gonesem/nes/ppu"
	"image"
	"image/color"
)

type NES struct {
	cpu       *cpu.CPU
	ppu       *ppu.PPU
	cartridge *cartridge.Cartridge

	ram [2048]uint8

	TotalCycles uint64

	dmaPage uint8
	dmaAddr uint8
	dmaData uint8

	dmaTransfer bool
	dmaDummy    bool
}

func NewNES(cartridge *cartridge.Cartridge, colorPalette [64]color.RGBA) *NES {
	nes := &NES{cartridge: cartridge, TotalCycles: 0}

	cpu := cpu.NewCPU(nes)
	ppu := ppu.NewPPU(cartridge, colorPalette)

	nes.cpu = cpu
	nes.ppu = ppu

	return nes
}

func (nes *NES) Read(addr uint16) uint8 {
	switch {
	case addr <= 0x1FFF:
		return nes.ram[addr%0x0800]
	case addr >= 0x2000 && addr <= 0x3FFF:
		return nes.ppu.CPURead(addr % 0x0008)
	default:
		return nes.cartridge.PRGRead(addr)
	}
}

func (nes *NES) Write(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		nes.ram[addr%0x0800] = value
	case addr >= 0x2000 && addr <= 0x3FFF:
		nes.ppu.CPUWrite(addr%0x0008, value)
	case addr == 0x4014:
		nes.dmaPage = value
		nes.dmaAddr = 0x00
		nes.dmaTransfer = true
	default:
		nes.cartridge.PRGWrite(addr, value)
	}
}

func (nes *NES) Clock() {
	nes.ppu.Clock()

	if nes.TotalCycles%3 == 0 {
		if nes.dmaTransfer {
			if nes.dmaDummy {
				if nes.TotalCycles%2 == 1 {
					nes.dmaDummy = false
				}
			} else {
				if nes.TotalCycles%2 == 0 {
					addr := uint16(nes.dmaPage)<<8 | uint16(nes.dmaAddr)
					nes.dmaData = nes.Read(addr)
				} else {
					nes.ppu.TransferDMAData(nes.dmaAddr, nes.dmaData)

					if nes.dmaData == 0x00 {
						nes.dmaTransfer = false
						nes.dmaDummy = true
					}
				}
			}
		} else {
			nes.cpu.Clock()
		}
	}

	if nes.ppu.EmitNMI {
		nes.cpu.NMI()
		nes.ppu.EmitNMI = false
	}

	nes.TotalCycles++
}

func (nes *NES) FrameComplete() bool {
	return nes.ppu.FrameComplete
}

func (nes *NES) ResetFrameComplete() {
	nes.ppu.FrameComplete = false
}

func (nes *NES) GetFrame() *image.RGBA {
	return nes.ppu.GetFrame()
}

func (nes *NES) GetPatternTable(tableIndex uint8, paletteIndex uint8) *image.RGBA {
	return nes.ppu.GetPatternTable(tableIndex, paletteIndex)
}
