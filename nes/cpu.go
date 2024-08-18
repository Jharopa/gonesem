package nes

import (
	"fmt"
)

type Status uint8

const (
	StatusCarry     Status = 1 << iota // C
	StatusZero                         // Z
	StatusInterrupt                    // I
	StatusDecimal                      // D
	StatusBreak                        // B
	StatusUnused                       // U
	StatusOverflow                     // V
	StatusNegative                     // N
)

const (
	StackPage  uint16 = 0x0100
	StackReset uint8  = 0x00FD
)

type AddressingMode uint8

const (
	AddressingModeImplied     AddressingMode = iota // IMP 0
	AddressingModeAccumulator                       // ACC 1
	AddressingModeImmediate                         // IMM 2
	AddressingModeZeroPage                          // ZP0 3
	AddressingModeZeroPageX                         // ZPX 4
	AddressingModeZeroPageY                         // ZPY 5
	AddressingModeRelative                          // REL 6
	AddressingModeAbsolute                          // ABS 7
	AddressingModeAbsoluteX                         // ABX 8
	AddressingModeAbsoluteY                         // ABY 9
	AddressingModeIndirect                          // IND 10
	AddressingModeIndirectX                         // IZX 11
	AddressingModeIndirectY                         // IZY 12
)

type InstructionArgs struct {
	addrMode AddressingMode
	address  uint16
}

type CPU struct {
	A  uint8  // Accumulator register
	X  uint8  // X index register
	Y  uint8  // Y index register
	PC uint16 // Program counter register
	SP uint8  // Statck pointer register
	SR Status // Status register

	RAM [65536]uint8
}

func NewCPU() *CPU {
	// 6502 registers at powerup
	cpu := &CPU{}
	cpu.Reset()

	return cpu
}

func (cpu *CPU) Reset() {
	// 6502 registers at reset
	cpu.A = 0
	cpu.X = 0
	cpu.Y = 0
	cpu.PC = cpu.ReadWord(0xFFFC)
	cpu.SP = StackReset
	cpu.SR = StatusUnused | StatusInterrupt
}

// ------------------------- //
// Memory read/write methods //
// ------------------------- //

/**
The 6502 has a maximum addressable space of 64KB (65536 unique 1-byte addresses)
and uses little-endian order which can be seen being enfornced in the in the
ReadWord and WriteWord helper methods.

Example of little-endian storing 16-bit value:
Value: 0xFF29
|----------------------------|
| Address | Value at address |
| 0x0200  | 0x29             |
| 0x0201  | 0xFF             |
|----------------------------|

Example of big-endian storing 16-bit for comparison:
Value: 0xFF29
|----------------------------|
| Address | Value at address |
| 0x0200  | 0xFF             |
| 0x0201  | 0x29             |
|----------------------------|
**/

// Returns value from memory at address addr
func (cpu *CPU) Read(addr uint16) uint8 {
	return cpu.RAM[addr]
}

// Writes value to address addr
func (cpu *CPU) Write(addr uint16, value uint8) {
	cpu.RAM[addr] = value
}

// Returns 16 bit value from memory at address addr converting from little-endian order
func (cpu *CPU) ReadWord(addr uint16) uint16 {
	lo := uint16(cpu.Read(addr))
	hi := uint16(cpu.Read(addr + 1))
	return (hi << 8) | lo
}

// Writes 16 bit value to address addr converting value to little-endian order
func (cpu *CPU) WriteWord(addr uint16, value uint16) {
	hi := uint8(value >> 8)
	lo := uint8(value & 0x00FF)
	cpu.Write(addr, lo)
	cpu.Write(addr+1, hi)
}

// ---------------------- //
// Stack push/pop methods //
// ---------------------- //

/**
* The 6502 stack implementation uses the page starting from memory address 0x0100.
* Upon power up the stack pointer starts at address 0xFD (absoulte address 0x01FD).
* As values are pushed onto the stack the stack pointer moves down the addressable space
  e.g. if a value was pushed on powerup, an 8-bit value would be pushed to address 0xFD
  and the stack pointer would be decremented to the next address 0xFC.
* Similarly to pushing, popping is the same process but in reverse order
  e.g. if the stack pointer is at 0xF0 then the stack pointer is incremented to 0xF1 and
  the value at this address is returned.
* As the 6502 uses little-endian order, when pushing or popping 16-bit values requires
  converting to or from little-endian, however as the stack moves down the address space
  from higher -> lower, the higher 8-bits of the 16-bit value are pushed to the higher
  address and the lower 8-bits a pushed to the lower address, the inverse of when reading
  or writing 16-bit values to other parts of memory.

Example of 16-bit value pushed to stack at powerup in little-endian order:

Value: 0xFF29
|----------------------------|
| Address | Value at address |
| 0x01FB  | 0x00             | <--- Stack Pointer
| 0x01FC  | 0xFF             |
| 0x01FD  | 0x29             |
|----------------------------|
**/

// Pushes 8-bit value onto the stack
func (cpu *CPU) Push(value uint8) {
	cpu.Write(StackPage+uint16(cpu.SP), value)
	cpu.SP--
}

// Pops 8-bit value from the stack returning the value
func (cpu *CPU) Pop() uint8 {
	cpu.SP++
	return cpu.Read(StackPage + uint16(cpu.SP))
}

// Pushes 16-bit value onto the stack converting it to little-endian order
func (cpu *CPU) PushWord(value uint16) {
	hi := uint8(value >> 8)
	lo := uint8(value & 0x00FF)
	cpu.Push(hi)
	cpu.Push(lo)
}

// Pops 16-bit value off the stack converting it from little-endian order before returning
func (cpu *CPU) PopWord() uint16 {
	lo := uint16(cpu.Pop())
	hi := uint16(cpu.Pop())
	return hi<<8 | lo
}

func (cpu *CPU) PrintRegisters() {
	fmt.Printf(
		"Accumulator register: %d\nIndex register X: %d\nIndex register Y: %d\nProgram counter: 0x%04X\nStack pointer: 0x%02X\n",
		cpu.A, cpu.X, cpu.Y, cpu.PC, cpu.SP,
	)

	cpu.PrintProcessorStatus()
}

func (cpu *CPU) PrintProcessorStatus() {
	fmt.Printf("NVUBDIZC\n")
	fmt.Printf("%08b\n", cpu.SR)
}

func (cpu *CPU) SetStatus(status Status, value bool) {
	if value {
		cpu.SR |= status
	} else {
		cpu.SR &^= status
	}
}

func (cpu *CPU) GetStatus(status Status) bool {
	return cpu.SR&status != 0
}

func (cpu *CPU) SetZ(value uint8) {
	if value == 0 {
		cpu.SetStatus(StatusZero, true)
	} else {
		cpu.SetStatus(StatusZero, false)
	}
}

func (cpu *CPU) SetN(value uint8) {
	if value&0x80 != 0 {
		cpu.SetStatus(StatusNegative, true)
	} else {
		cpu.SetStatus(StatusNegative, false)
	}
}

func (cpu *CPU) SetZN(value uint8) {
	cpu.SetZ(value)
	cpu.SetN(value)
}

func (cpu *CPU) Clock() {

}

func (cpu *CPU) FetchOperandAddress(addrMode AddressingMode) uint16 {
	switch addrMode {
	// Instruction's operand is implict to the intrustion or does not exist.
	case AddressingModeImplied:
		return 0

	// Instruction's operand is within the accumulator.
	case AddressingModeAccumulator:
		return 0

	// Address of instructions operand is immediately adjacent to the current program counter address.
	case AddressingModeImmediate:
		return cpu.PC + 1

	// Address of instructions operand is the 8-bit value in the address immediately adjacent to
	// the current program counter address mapped to the 0th page.
	case AddressingModeZeroPage:
		return uint16(cpu.Read(cpu.PC + 1))

	// The same as zero page address but with X registers value applied as offset
	// Mask applied to the 8 MSB in case of overflow due to offset.
	case AddressingModeZeroPageX:
		return uint16(cpu.Read(cpu.PC+1)+cpu.X) & 0x00FF

	// The same as zero page address but with Y registers value applied as offset
	// Mask applied to the 8 MSB in case of overflow due to offset.
	case AddressingModeZeroPageY:
		return uint16(cpu.Read(cpu.PC+1)+cpu.Y) & 0x00FF

	// Special addressing mode for branching operations, the 8-bit value found at the address
	// directly adjacent to the instructions operator is treated as a signed 2's compliment number
	// (-128 to 127). This offset is then applied to the address following that to jump to an address
	// within a 128 range around the address of the inital 8-bit value.
	case AddressingModeRelative:
		relAddr := uint16(cpu.Read(cpu.PC + 1))

		if relAddr&0x80 != 0 {
			relAddr |= 0xFF00
		}

		return (cpu.PC + 2) + relAddr

	// Reads the 2 bytes starting from the address directly adjacent to the operator and treats
	// that 16-bit number as an absoulte address.
	case AddressingModeAbsolute:
		return cpu.ReadWord(cpu.PC + 1)

	// Same as absolute but the 16-bit address has the contents of the X register applied to it
	// as an offset.
	case AddressingModeAbsoluteX:
		return cpu.ReadWord(cpu.PC+1) + uint16(cpu.X)

	// Same as absolute but the 16-bit address has the contents of the Y register applied to it
	// as an offset.
	case AddressingModeAbsoluteY:
		return cpu.ReadWord(cpu.PC+1) + uint16(cpu.Y)

	/**
		6502's implementation of pointers, 16-bit address read from address directly after opcode,
		the value at this address is the address where the operand is acutally stored

		NOTE. This address mode has a hardware bug on the 6502, if the pointer address
		is on the last address of the page instead of the high byte of the address being read
		from the 0th address of the next page the high byte wraps around to the 0th address of the
		page the low byte is on, for example:

		|----------------------------|
		| Address | Value at address |
		| 0x0200  | 0xFF             | <--- 3. Instead the read wraps around to the 0th address of
		| ......  | ....             |      the low bytes current page making the operands address
		| ......  | ....             |      0xFF29.
		| ......  | ....             |
		| 0x02FD  | 0x00             |
		| 0x02FE  | 0x00             |
		| 0x02FF  | 0x29             | <--- 1. Low byte of ptr address residing on page boundary.
		| 0x0300  | 0x11             | <--- 2. Expected that high by will be read 0th address on
		|----------------------------|      following page making operands address 0x1129.
	**/
	case AddressingModeIndirect:
		ptrAddr := cpu.ReadWord(cpu.PC + 1)

		if ptrAddr&0x00FF == 0x00FF {
			return uint16(cpu.Read(ptrAddr&0xFF00))<<8 | uint16(cpu.Read(ptrAddr))
		} else {
			return cpu.ReadWord(ptrAddr)
		}

	case AddressingModeIndirectX:
		ptr := uint16(cpu.Read(cpu.PC + 1))

		lo := uint16(cpu.Read((ptr + uint16(cpu.X)) & 0x00FF))
		hi := uint16(cpu.Read((ptr + uint16(cpu.X) + 1) & 0x00FF))

		return hi<<8 | lo

	case AddressingModeIndirectY:
		ptr := uint16(cpu.Read(cpu.PC + 1))

		lo := uint16(cpu.Read(ptr) & 0x00FF)
		hi := uint16(cpu.Read(ptr+1) & 0x00FF)

		return (hi<<8 | lo) + uint16(cpu.Y)

	default:
		panic(fmt.Sprintf("Invalid addressing mode %d", addrMode))
	}
}

/*
*
Add with carry
*
*/
func (cpu *CPU) ADC(args InstructionArgs) {

}

/*
*
Logical And
* Logical And operand with contents of accumulator
* Set zero status if resulting value is 0
* Set negative status if resulting value's 7th bit is set
*
*/
func (cpu *CPU) ADD(args InstructionArgs) {
	operand := cpu.Read(args.address)
	cpu.A &= operand
	cpu.SetZN(cpu.A)
}

/*
*
Arithmetic Shift Left
* Shift contents of address 1 bit left
* Set contents of bit 7 in carry status
* Set zero status if resulting value is 0
* Set negative status if resulting value's 7th bit is set

Function contains two paths that performs above steps, one working
on the accumulator and the other on a memory address depending on the
addressing mode of instruction.
*
*/
func (cpu *CPU) ASL(args InstructionArgs) {
	if args.addrMode == AddressingModeAccumulator {
		cpu.SetStatus(StatusCarry, cpu.A&0x80 != 0)
		cpu.A <<= 1
		cpu.SetZN(cpu.A)
	} else {
		operand := cpu.Read(args.address)
		cpu.SetStatus(StatusCarry, operand&0x80 != 0)
		operand <<= 1
		cpu.SetZN(operand)
		cpu.Write(args.address, operand)
	}
}

/*
*
Branch if Carry Clear
* Checks if carry bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BCC(args InstructionArgs) {
	if !cpu.GetStatus(StatusCarry) {
		cpu.PC = args.address
	}
}

/*
*
Branch if Carry Set
* Checks if carry bit in status register is set, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BCS(args InstructionArgs) {
	if cpu.GetStatus(StatusCarry) {
		cpu.PC = args.address
	}
}

/*
*
Branch if Equal
* Checks if zero bit in status register is set, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BEQ(args InstructionArgs) {
	if cpu.GetStatus(StatusZero) {
		cpu.PC = args.address
	}
}

/*
*
Bit Test
* Reads operand from memory
* ANDs the contents of the accumulator with the operand from memory, setting
the zero status flag based on the result of that operation.
* Values of bit 6 and 7 of operand are used to set the negative and overflow
status flags respectively.
*
*/
func (cpu *CPU) BIT(args InstructionArgs) {
	operand := cpu.Read(args.address)

	cpu.SetZ(cpu.A & operand)
	cpu.SetStatus(StatusOverflow, operand&(1<<6) != 0)
	cpu.SetStatus(StatusNegative, operand&(1<<7) != 0)
}

/*
*
Branch if Minus
* Checks if negative bit in status register is set, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BMI(args InstructionArgs) {
	if cpu.GetStatus(StatusNegative) {
		cpu.PC = args.address
	}
}

/*
*
Branch if Not Equal
* Checks if zero bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BNE(args InstructionArgs) {
	if !cpu.GetStatus(StatusZero) {
		cpu.PC = args.address
	}
}

/*
*
Branch if Positive
* Checks if negative bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BPL(args InstructionArgs) {
	if !cpu.GetStatus(StatusNegative) {
		cpu.PC = args.address
	}
}

func (cpu *CPU) BRK(args InstructionArgs) {

}

/*
*
Branch if Overflow is Clear
* Checks if overflow bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BVC(args InstructionArgs) {
	if !cpu.GetStatus(StatusOverflow) {
		cpu.PC = args.address
	}
}

/*
*
Branch if Overflow is Set
* Checks if overflow bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func (cpu *CPU) BVS(args InstructionArgs) {
	if cpu.GetStatus(StatusOverflow) {
		cpu.PC = args.address
	}
}
