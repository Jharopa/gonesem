package cpu

import (
	"fmt"
	"gonesem/nes/util"
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

type OperationArgs struct {
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

	cycles uint8 // Cycles remaining for current instruction execution

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

	cpu.cycles = 0
}

func (cpu *CPU) Clock() bool {
	if cpu.cycles > 0 {
		cpu.cycles--
		return cpu.cycles == 0
	}

	opcode := cpu.Read(cpu.PC)

	instruction := instructions[opcode]

	args := OperationArgs{
		instruction.AddressingMode,
		cpu.fetchOperandAddress(instruction.AddressingMode),
	}

	cpu.cycles = instruction.InstructionCycles - 1
	cpu.PC += uint16(instruction.InstructionSize)

	instruction.operation(cpu, args)

	return false
}

func (cpu *CPU) fetchOperandAddress(addrMode AddressingMode) uint16 {
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

func (cpu *CPU) PrintCPUState() {
	cpu.PrintRegisters()
	cpu.PrintProcessorStatus()
}

func (cpu *CPU) PrintRegisters() {
	fmt.Printf(
		"Accumulator register: %d\nIndex register X: %d\nIndex register Y: %d\nProgram counter: 0x%04X\nStack pointer: 0x%02X\n",
		cpu.A, cpu.X, cpu.Y, cpu.PC, cpu.SP,
	)
}

func (cpu *CPU) PrintProcessorStatus() {
	fmt.Printf("NVUBDIZC\n")
	fmt.Printf("%08b\n", cpu.SR)
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
func (cpu *CPU) push(value uint8) {
	cpu.Write(StackPage+uint16(cpu.SP), value)
	cpu.SP--
}

// Pops 8-bit value from the stack returning the value
func (cpu *CPU) pop() uint8 {
	cpu.SP++
	return cpu.Read(StackPage + uint16(cpu.SP))
}

// Pushes 16-bit value onto the stack converting it to little-endian order
func (cpu *CPU) pushWord(value uint16) {
	hi := uint8(value >> 8)
	lo := uint8(value & 0x00FF)
	cpu.push(hi)
	cpu.push(lo)
}

// Pops 16-bit value off the stack converting it from little-endian order before returning
func (cpu *CPU) popWord() uint16 {
	lo := uint16(cpu.pop())
	hi := uint16(cpu.pop())
	return hi<<8 | lo
}

func (cpu *CPU) setStatus(status Status, value bool) {
	if value {
		cpu.SR |= status
	} else {
		cpu.SR &^= status
	}
}

func (cpu *CPU) getStatus(status Status) bool {
	return cpu.SR&status != 0
}

func (cpu *CPU) setZ(value uint8) {
	if value == 0 {
		cpu.setStatus(StatusZero, true)
	} else {
		cpu.setStatus(StatusZero, false)
	}
}

func (cpu *CPU) setN(value uint8) {
	if value&0x80 != 0 {
		cpu.setStatus(StatusNegative, true)
	} else {
		cpu.setStatus(StatusNegative, false)
	}
}

func (cpu *CPU) setZN(value uint8) {
	cpu.setZ(value)
	cpu.setN(value)
}

/*
*
Add with carry
*
*/
func adc(cpu *CPU, args OperationArgs) {
	operand := uint16(cpu.Read(args.address))
	carryBit := uint16(util.Btou8(cpu.getStatus(StatusCarry)))

	result := uint16(cpu.A) + operand + carryBit

	overflowed := ((uint16(cpu.A) ^ result) & ^(uint16(cpu.A) ^ operand) & 0x0080) != 0

	cpu.setStatus(StatusOverflow, overflowed)
	cpu.setStatus(StatusCarry, result > 255)
	cpu.setZN(uint8(result))

	cpu.A = uint8(result)
}

/*
*
Logical And
* Logical And operand with contents of accumulator
* Set zero status if resulting value is 0
* Set negative status if resulting value's 7th bit is set
*
*/
func and(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.Read(args.address)
	cpu.setZN(cpu.A)
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
func asl(cpu *CPU, args OperationArgs) {
	if args.addrMode == AddressingModeAccumulator {
		cpu.setStatus(StatusCarry, cpu.A&0x80 != 0)
		cpu.A <<= 1
		cpu.setZN(cpu.A)
	} else {
		operand := cpu.Read(args.address)
		cpu.setStatus(StatusCarry, operand&0x80 != 0)
		operand <<= 1
		cpu.setZN(operand)
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
func bcc(cpu *CPU, args OperationArgs) {
	if !cpu.getStatus(StatusCarry) {
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
func bcs(cpu *CPU, args OperationArgs) {
	if cpu.getStatus(StatusCarry) {
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
func beq(cpu *CPU, args OperationArgs) {
	if cpu.getStatus(StatusZero) {
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
func bit(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address)

	cpu.setZ(cpu.A & operand)
	cpu.setStatus(StatusOverflow, operand&(1<<6) != 0)
	cpu.setStatus(StatusNegative, operand&(1<<7) != 0)
}

/*
*
Branch if Minus
* Checks if negative bit in status register is set, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func bmi(cpu *CPU, args OperationArgs) {
	if cpu.getStatus(StatusNegative) {
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
func bne(cpu *CPU, args OperationArgs) {
	if !cpu.getStatus(StatusZero) {
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
func bpl(cpu *CPU, args OperationArgs) {
	if !cpu.getStatus(StatusNegative) {
		cpu.PC = args.address
	}
}

func brk(cpu *CPU, args OperationArgs) {
	cpu.pushWord(cpu.PC)
	cpu.push(uint8(cpu.SR | StatusBreak | StatusUnused))
	cpu.setStatus(StatusInterrupt, true)
	cpu.PC = cpu.ReadWord(0xFFFE)
}

/*
*
Branch if Overflow is Clear
* Checks if overflow bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func bvc(cpu *CPU, args OperationArgs) {
	if !cpu.getStatus(StatusOverflow) {
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
func bvs(cpu *CPU, args OperationArgs) {
	if cpu.getStatus(StatusOverflow) {
		cpu.PC = args.address
	}
}

func clc(cpu *CPU, args OperationArgs) {
	cpu.setStatus(StatusCarry, false)
}

func cld(cpu *CPU, args OperationArgs) {
	cpu.setStatus(StatusDecimal, false)
}

func cli(cpu *CPU, args OperationArgs) {
	cpu.setStatus(StatusInterrupt, false)
}

func clv(cpu *CPU, args OperationArgs) {
	cpu.setStatus(StatusOverflow, false)
}

/*
*
Compare
* Reads operand from memory
* Sets carry bit of status regsiter if accumulator's contents is >= operand
* Sets zero bit of status register if accumulators's contents == operand
* Sets negative bit of status register if accumulators's contents < operand
*
*/
func cmp(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address)

	cpu.setStatus(StatusCarry, cpu.A >= operand)
	cpu.setZN(cpu.A - operand)
}

/*
*
Compare X Register
* Reads operand from memory
* Sets carry bit of status regsiter if X registers' contents is >= operand
* Sets zero bit of status register if X registers' contents == operand
* Sets negative bit of status register if X registers' contents < operand
*
*/
func cpx(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address)

	cpu.setStatus(StatusCarry, cpu.X >= operand)
	cpu.setZN(cpu.X - operand)
}

/*
*
Compare Y Register
* Reads operand from memory
* Sets carry bit of status regsiter if Y registers' contents is >= operand
* Sets zero bit of status register if Y registers' contents == operand
* Sets negative bit of status register if Y registers' contents < operand
*
*/
func cpy(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address)

	cpu.setStatus(StatusCarry, cpu.Y >= operand)
	cpu.setZN(cpu.Y - operand)
}

func dec(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address) - 1

	cpu.Write(args.address, operand)
	cpu.setZN(operand)
}

func dex(cpu *CPU, args OperationArgs) {
	cpu.X--
	cpu.setZN(cpu.X)
}

func dey(cpu *CPU, args OperationArgs) {
	cpu.Y--
	cpu.setZN(cpu.Y)
}

func eor(cpu *CPU, args OperationArgs) {
	cpu.A ^= cpu.Read(args.address)
	cpu.setZN(cpu.A)
}

func inc(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address) + 1

	cpu.Write(args.address, operand)
	cpu.setZN(operand)
}

func inx(cpu *CPU, args OperationArgs) {
	cpu.X++
	cpu.setZN(cpu.X)
}

func iny(cpu *CPU, args OperationArgs) {
	cpu.Y++
	cpu.setZN(cpu.Y)
}

func jmp(cpu *CPU, args OperationArgs) {
	cpu.pushWord(cpu.PC - 1)
	cpu.PC = args.address
}

func jsr(cpu *CPU, args OperationArgs) {
	cpu.pushWord(cpu.PC - 1)
	cpu.PC = args.address
}

func lda(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.Read(args.address)
	cpu.setZN(cpu.A)
}

func ldx(cpu *CPU, args OperationArgs) {
	cpu.X = cpu.Read(args.address)
	cpu.setZN(cpu.X)
}

func ldy(cpu *CPU, args OperationArgs) {
	cpu.Y = cpu.Read(args.address)
	cpu.setZN(cpu.Y)
}

func lsr(cpu *CPU, args OperationArgs) {
	if args.addrMode == AddressingModeAccumulator {
		cpu.setStatus(StatusCarry, cpu.A&0x0001 != 0)
		cpu.A >>= 1
		cpu.setZN(cpu.A)
	} else {
		operand := cpu.Read(args.address)
		cpu.setStatus(StatusCarry, operand&0x0001 != 0)
		operand >>= 1
		cpu.setZN(operand)
		cpu.Write(args.address, operand)
	}
}

func nop(cpu *CPU, args OperationArgs) {
}

func ora(cpu *CPU, args OperationArgs) {
	cpu.A |= cpu.Read(args.address)
	cpu.setZN(cpu.A)
}

func pha(cpu *CPU, args OperationArgs) {
	cpu.push(cpu.A)
}

func php(cpu *CPU, args OperationArgs) {
	cpu.push(uint8(cpu.SR | StatusBreak | StatusUnused))
}

func pla(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.pop()
	cpu.setZN(cpu.A)
}

func plp(cpu *CPU, args OperationArgs) {
	cpu.SR = Status(cpu.pop())
	cpu.setStatus(StatusUnused, true)
}

func rol(cpu *CPU, args OperationArgs) {
	carryBit := util.Btou8(cpu.getStatus(StatusCarry))

	if args.addrMode == AddressingModeAccumulator {
		cpu.setStatus(StatusCarry, cpu.A&0x80 != 0)
		cpu.A = cpu.A<<1 | carryBit
		cpu.setZN(cpu.A)
	} else {
		operand := cpu.Read(args.address)
		cpu.setStatus(StatusCarry, operand&0x80 != 0)
		operand = operand<<1 | carryBit
		cpu.setZN(operand)
		cpu.Write(args.address, operand)
	}
}

func ror(cpu *CPU, args OperationArgs) {
	carryBit := util.Btou8(cpu.getStatus(StatusCarry)) << 7

	if args.addrMode == AddressingModeAccumulator {
		cpu.setStatus(StatusCarry, cpu.A&0x0001 != 0)
		cpu.A = cpu.A>>1 | carryBit
		cpu.setZN(cpu.A)
	} else {
		operand := cpu.Read(args.address)
		cpu.setStatus(StatusCarry, operand&0x0001 != 0)
		operand = operand>>1 | carryBit
		cpu.setZN(operand)
		cpu.Write(args.address, operand)
	}
}

func rti(cpu *CPU, args OperationArgs) {
	cpu.SR = Status(cpu.pop())
	cpu.setStatus(StatusBreak, false)
	cpu.setStatus(StatusUnused, true)

	cpu.PC = cpu.popWord()
}

func rts(cpu *CPU, args OperationArgs) {
	cpu.PC = cpu.popWord() + 1
}

func sbc(cpu *CPU, args OperationArgs) {
	operand := uint16(cpu.Read(args.address)) ^ 0x00FF
	carryBit := uint16(util.Btou8(cpu.getStatus(StatusCarry)))

	result := uint16(cpu.A) + operand + carryBit

	overflowed := ((uint16(cpu.A) ^ result) & ^(uint16(cpu.A) ^ operand) & 0x0080) != 0

	cpu.setStatus(StatusOverflow, overflowed)
	cpu.setStatus(StatusCarry, result > 255)
	cpu.setZN(uint8(result))

	cpu.A = uint8(result)
}

func sec(cpu *CPU, args OperationArgs) {
	cpu.setStatus(StatusCarry, true)
}

func sed(cpu *CPU, args OperationArgs) {
	cpu.setStatus(StatusDecimal, true)
}

func sei(cpu *CPU, args OperationArgs) {
	cpu.setStatus(StatusInterrupt, true)
}

func sta(cpu *CPU, args OperationArgs) {
	cpu.Write(args.address, cpu.A)
}

func stx(cpu *CPU, args OperationArgs) {
	cpu.Write(args.address, cpu.X)
}

func sty(cpu *CPU, args OperationArgs) {
	cpu.Write(args.address, cpu.Y)
}

func tax(cpu *CPU, args OperationArgs) {
	cpu.X = cpu.A
	cpu.setZN(cpu.X)
}

func tay(cpu *CPU, args OperationArgs) {
	cpu.Y = cpu.A
	cpu.setZN(cpu.Y)
}

func tsx(cpu *CPU, args OperationArgs) {
	cpu.X = cpu.SP
	cpu.setZN(cpu.Y)
}

func txa(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.X
	cpu.setZN(cpu.A)
}

func txs(cpu *CPU, args OperationArgs) {
	cpu.SP = cpu.X
}

func tya(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.Y
	cpu.setZN(cpu.A)
}

func xxx(cpu *CPU, args OperationArgs) {

}

// ----------------- //
// Unoffical Opcodes //
// ----------------- //

/**
Load Accumulator Logical Shift Right
* Performs equvilant of a LDA followed by an LSR
**/
func alr(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.Read(args.address)
	cpu.setStatus(StatusCarry, cpu.A&0x0001 != 0)
	cpu.A >>= 1
	cpu.setZN(cpu.A)
}

func anc(cpu *CPU, arg OperationArgs) {
	
}
