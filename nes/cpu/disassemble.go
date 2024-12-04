package cpu

import (
	"fmt"
	"strings"
)

func (cpu *CPU) NintendulatorDisassembly() string {
	var sb strings.Builder

	opcode := cpu.Read(cpu.PC)

	instruction := Instructions[opcode]

	instructionSize := instruction.Size

	sb.WriteString(fmt.Sprintf("%04X  ", cpu.PC))

	for i := 0; i < int(instructionSize); i++ {
		address := cpu.PC + uint16(i)
		sb.WriteString(fmt.Sprintf("%02X ", cpu.Read(address)))
	}

	sb.WriteString(strings.Repeat(" ", 16-sb.Len()))

	disassembledInstruction := cpu.DisassembleCPUInstruction()
	sb.WriteString(disassembledInstruction)

	sb.WriteString(strings.Repeat(" ", 47-sb.Len()))

	sb.WriteString(fmt.Sprintf(" A:%02X", cpu.A))
	sb.WriteString(fmt.Sprintf(" X:%02X", cpu.X))
	sb.WriteString(fmt.Sprintf(" Y:%02X", cpu.Y))
	sb.WriteString(fmt.Sprintf(" P:%02X", cpu.SR))
	sb.WriteString(fmt.Sprintf(" SP:%02X", cpu.SP))

	sb.WriteString(fmt.Sprintf(" CYC:%d", cpu.TotalCycles))

	return sb.String()
}

func (cpu *CPU) DisassembleCPUInstruction() string {
	opcode := cpu.Read(cpu.PC)

	instruction := Instructions[opcode]

	var sb strings.Builder
	var instructionArg uint16

	if instruction.Size == 2 {
		instructionArg = uint16(cpu.Read(cpu.PC + 1))
	} else if instruction.Size == 3 {
		instructionArg = cpu.ReadWord(cpu.PC + 1)
	}

	sb.WriteString(fmt.Sprintf("%s ", instruction.Mnemonic))

	switch instruction.AddressingMode {
	case AddressingModeImplied:
		break
	case AddressingModeAccumulator:
		sb.WriteString("A")
	case AddressingModeImmediate:
		sb.WriteString(fmt.Sprintf("#$%02X", instructionArg))
	case AddressingModeZeroPage:
		sb.WriteString(fmt.Sprintf("$%02X", instructionArg))
	case AddressingModeZeroPageX:
		sb.WriteString(fmt.Sprintf("$%02X,X", instructionArg))
	case AddressingModeZeroPageY:
		sb.WriteString(fmt.Sprintf("$%02X,Y", instructionArg))
	case AddressingModeRelative:
		if instructionArg&0x80 != 0 {
			instructionArg |= 0xFF00
		}

		sb.WriteString(fmt.Sprintf("$%02X", instructionArg+cpu.PC+2))
	case AddressingModeAbsolute:
		sb.WriteString(fmt.Sprintf("$%04X", instructionArg))
	case AddressingModeAbsoluteX:
		sb.WriteString(fmt.Sprintf("$%04X,X", instructionArg))
	case AddressingModeAbsoluteY:
		sb.WriteString(fmt.Sprintf("$%04X,Y", instructionArg))
	case AddressingModeIndirect:
		sb.WriteString(fmt.Sprintf("($%04X)", instructionArg))
	case AddressingModeIndirectX:
		sb.WriteString(fmt.Sprintf("($%02X,X)", instructionArg))
	case AddressingModeIndirectY:
		sb.WriteString(fmt.Sprintf("($%02X),Y", instructionArg))
	}

	return sb.String()
}
