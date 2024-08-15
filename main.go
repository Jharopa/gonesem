package main

import (
	"nesem/nes"
)

func main() {
	cpu := nes.NewCPU()
	
	cpu.PrintProcessorStatus()

	cpu.SetStatus(nes.StatusInterrupt, false)

	cpu.PrintProcessorStatus()

	cpu.SetStatus(nes.StatusOverflow, true)

	cpu.PrintProcessorStatus()

	cpu.SetStatus(nes.StatusOverflow, false)

	cpu.PrintProcessorStatus()
}