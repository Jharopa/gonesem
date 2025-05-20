package nes_test

import (
	"bufio"
	"io"
	"log"
	"os"
	"testing"

	"gonesem/internal/cpu"
)

type TestMemory struct {
	RAM [65535]uint8
}

func (memory *TestMemory) Read(addr uint16) uint8 {
	return memory.RAM[addr]
}

func (memory *TestMemory) Write(addr uint16, value uint8) {
	memory.RAM[addr] = value
}

func loadNestest() []byte {
	file, err := os.Open("./data/nestest.nes")

	if err != nil {
		log.Printf("Failed to open netstest.nes file: %s", err)
		os.Exit(1)
	}

	stat, err := file.Stat()

	if err != nil {
		log.Printf("Failed to retrieve netstest.nes file stats: %s", err)
		os.Exit(1)
	}

	rom := make([]byte, stat.Size())

	_, err = bufio.NewReader(file).Read(rom)

	if err != nil && err != io.EOF {
		log.Printf("Failed to read file into rom buffer: %s", err)
		os.Exit(1)
	}

	return rom
}

func TestNestest(t *testing.T) {
	rom := loadNestest()

	memory := &TestMemory{}

	copy(memory.RAM[0xC000:0xFFFF], rom[0x10:0x4000])

	testCPU := cpu.NewCPU(memory)

	testCPU.PC = 0xC000

	for {
		complete := false

		for !complete {
			complete = testCPU.Clock()
		}

		if testCPU.Read(0x0002) != 0x00 {
			testCPU.PrintCPUState(false)
			t.Fatalf("Official instruction failed: 0%02Xh", testCPU.Read(0x0002))
		}

		if testCPU.Read(0x0003) != 0x00 {
			testCPU.PrintCPUState(false)
			t.Fatalf("Unofficial instruction failed: 0%02Xh", testCPU.Read(0x0003))
		}

		if testCPU.PC == 0xC66E {
			break
		}
	}
}

func TestNestestNintendulatorLog(t *testing.T) {
	rom := loadNestest()

	memory := &TestMemory{}

	copy(memory.RAM[0xC000:0xFFFF], rom[0x10:0x4000])

	testCPU := cpu.NewCPU(memory)

	testCPU.PC = 0xC000

	file, err := os.Open("./data/nestest.log")

	if err != nil {
		log.Printf("Failed to open netstest.log file: %s", err)
		os.Exit(1)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for {
		if !scanner.Scan() {
			break
		}

		expected := scanner.Text()
		actual := testCPU.NintendulatorDisassembly()

		if expected != actual {
			t.Fatalf("CPU disassembly did not match nestest.log\n Expected:\t%s\n Actual:\t%s\n", expected, actual)
		}

		complete := false

		for !complete {
			complete = testCPU.Clock()
		}
	}
}
