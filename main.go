package main

import (
	nes "gonesem/nes/cpu"
)

func main() {
	cpu := nes.NewCPU()

	cpu.PrintRegisters()

	complete := false

	for !complete {
		complete = cpu.Clock()
		cpu.PrintRegisters()
	}
}
