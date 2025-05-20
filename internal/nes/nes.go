package nes

import (
	"gonesem/internal/cartridge"
	"gonesem/internal/cpu"
	"gonesem/internal/ppu"
	"image"
	"image/color"
)

type NES struct {
	cpu       *cpu.CPU
	ppu       *ppu.PPU
	cartridge *cartridge.Cartridge

	ram [2048]uint8

	Controller      [2]uint8
	controllerState [2]uint8

	dmaPage uint8
	dmaAddr uint8
	dmaData uint8

	dmaTransfer bool
	dmaDummy    bool

	TotalCycles uint64
}

func NewNES(cartridge *cartridge.Cartridge, colorPalette [64]color.RGBA) *NES {
	nes := &NES{
		cartridge:   cartridge,
		dmaPage:     0,
		dmaAddr:     0,
		dmaData:     0,
		dmaTransfer: false,
		dmaDummy:    true,
		TotalCycles: 0,
	}

	nes.cpu = cpu.NewCPU(nes)
	nes.ppu = ppu.NewPPU(cartridge, colorPalette)

	return nes
}

func (nes *NES) Read(addr uint16) uint8 {
	var value uint8 = 0x00

	switch {
	case addr <= 0x1FFF:
		value = nes.ram[addr%0x0800]
	case addr >= 0x2000 && addr <= 0x3FFF:
		value = nes.ppu.CPURead(addr % 0x0008)
	case addr == 0x4016 || addr == 0x4017:
		if nes.controllerState[addr&0x0001]&0x80 > 0 {
			value = 1
		} else {
			value = 0
		}

		nes.controllerState[addr&0x0001] <<= 1
	default:
		value = nes.cartridge.PRGRead(addr)
	}

	return value
}

func (nes *NES) Write(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		nes.ram[addr&0x07FF] = value
	case addr >= 0x2000 && addr <= 0x3FFF:
		nes.ppu.CPUWrite(addr&0x0007, value)
	case addr == 0x4014:
		nes.dmaPage = value
		nes.dmaAddr = 0x00
		nes.dmaTransfer = true
	case addr == 0x4016 || addr == 0x4017:
		nes.controllerState[addr&0x0001] = nes.Controller[addr&0x0001]
	default:
		nes.cartridge.PRGWrite(addr, value)
	}
}

func (nes *NES) Clock() {
	nes.ppu.Clock()

	// The CPU clocks once for every three PPU clocks
	if nes.TotalCycles%3 == 0 {
		// While a DMA transfer is happening the CPU's clock is suspended.
		if nes.dmaTransfer {
			// DMA dummy syncs the DMA transfer with total cycles. This ensures the
			// OAM data read from the CPU during the DMA transfer occurs on even cycles and
			// the subsquent write of that data to PPU memory occurs on that next odd cycle.
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
					nes.dmaAddr++

					// When the incrementimg 8-bit DMA address wraps back around to 0
					// the full 256 bytes of OAM memory have been written to the
					// PPU memory and the DMA transfer has completed.
					if nes.dmaAddr == 0x00 {
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
