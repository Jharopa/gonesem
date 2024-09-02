package cpu_test

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"gonesem/nes/cpu"
)

func nintendulatorDisassemble(cpuPtr *cpu.CPU) string {
	var sb strings.Builder

	opcode := cpuPtr.Read(cpuPtr.PC)

	instruction := cpu.Instructions[opcode]

	instructionSize := instruction.Size

	sb.WriteString(fmt.Sprintf("%04X  ", cpuPtr.PC))

	for i := 0; i < int(instructionSize); i++ {
		address := cpuPtr.PC + uint16(i)
		sb.WriteString(fmt.Sprintf("%02X ", cpuPtr.Read(address)))
	}

	sb.WriteString(strings.Repeat(" ", 16-sb.Len()))

	disassembledInstruction := disassembleCPUInstruction(cpuPtr)
	sb.WriteString(disassembledInstruction)

	sb.WriteString(strings.Repeat(" ", 47-sb.Len()))

	sb.WriteString(fmt.Sprintf(" A:%02X", cpuPtr.A))
	sb.WriteString(fmt.Sprintf(" X:%02X", cpuPtr.X))
	sb.WriteString(fmt.Sprintf(" Y:%02X", cpuPtr.Y))
	sb.WriteString(fmt.Sprintf(" P:%02X", cpuPtr.SR))
	sb.WriteString(fmt.Sprintf(" SP:%02X", cpuPtr.SP))

	sb.WriteString(fmt.Sprintf(" CYC:%d", cpuPtr.TotalCycles))

	return sb.String()
}

func disassembleCPUInstruction(cpuPtr *cpu.CPU) string {
	opcode := cpuPtr.Read(cpuPtr.PC)

	instruction := cpu.Instructions[opcode]

	var sb strings.Builder
	var instructionArg uint16

	if instruction.Size == 2 {
		instructionArg = uint16(cpuPtr.Read(cpuPtr.PC + 1))
	} else if instruction.Size == 3 {
		instructionArg = cpuPtr.ReadWord(cpuPtr.PC + 1)
	}

	sb.WriteString(fmt.Sprintf("%s ", instruction.Mnemonic))

	switch instruction.AddressingMode {
	case cpu.AddressingModeImplied:
		break
	case cpu.AddressingModeAccumulator:
		sb.WriteString("A")
	case cpu.AddressingModeImmediate:
		sb.WriteString(fmt.Sprintf("#$%02X", instructionArg))
	case cpu.AddressingModeZeroPage:
		sb.WriteString(fmt.Sprintf("$%02X", instructionArg))
	case cpu.AddressingModeZeroPageX:
		sb.WriteString(fmt.Sprintf("$%02X,X", instructionArg))
	case cpu.AddressingModeZeroPageY:
		sb.WriteString(fmt.Sprintf("$%02X,Y", instructionArg))
	case cpu.AddressingModeRelative:
		if instructionArg&0x80 != 0 {
			instructionArg |= 0xFF00
		}

		sb.WriteString(fmt.Sprintf("$%02X", instructionArg+cpuPtr.PC+2))
	case cpu.AddressingModeAbsolute:
		sb.WriteString(fmt.Sprintf("$%04X", instructionArg))
	case cpu.AddressingModeAbsoluteX:
		sb.WriteString(fmt.Sprintf("$%04X,X", instructionArg))
	case cpu.AddressingModeAbsoluteY:
		sb.WriteString(fmt.Sprintf("$%04X,Y", instructionArg))
	case cpu.AddressingModeIndirect:
		sb.WriteString(fmt.Sprintf("($%04X)", instructionArg))
	case cpu.AddressingModeIndirectX:
		sb.WriteString(fmt.Sprintf("($%02X,X)", instructionArg))
	case cpu.AddressingModeIndirectY:
		sb.WriteString(fmt.Sprintf("($%02X),Y", instructionArg))
	}

	return sb.String()
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

	testCPU := cpu.NewCPU()

	testCPU.PC = 0xC000

	copy(testCPU.RAM[0xC000:0xFFFF], rom[0x10:0x4000])

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

	testCPU := cpu.NewCPU()

	testCPU.PC = 0xC000

	copy(testCPU.RAM[0xC000:0xFFFF], rom[0x10:0x4000])

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
		actual := nintendulatorDisassemble(testCPU)

		if expected != actual {
			t.Fatalf("CPU disassembly did not match nestest.log\n Expected:\t%s\n Actual:\t%s\n", expected, actual)
		}

		complete := false

		for !complete {
			complete = testCPU.Clock()
		}
	}
}
