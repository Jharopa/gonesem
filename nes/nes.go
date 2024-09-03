package nes

import "gonesem/nes/cpu"

type NES struct {
	CPU *cpu.CPU
	ram [2048]uint8
}

func NewNES() (*NES){
	nes := &NES{}

	cpu := cpu.NewCPU(nes)
	
	nes.CPU = cpu

	return nes
}

func (nes *NES) Read(addr uint16) uint8 {
	if addr <= 0x1FFF {
		return nes.ram[addr & 0x07FF]
	} else {
		return 0
	}
}

func (nes *NES) Write(addr uint16, value uint8) {
	if addr <= 0x1FFF {
		nes.ram[addr & 0x07FF] = value
	}
}