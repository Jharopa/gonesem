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

const (
	ResetVector uint16 = 0xFFFC
	IRQVector   uint16 = 0xFFFE
	NMIVector   uint16 = 0xFFFA
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

	cycles      uint8  // Cycles remaining for current instruction execution
	TotalCycles uint64 // Total instruction cycles over lifetime of CPU

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
	cpu.PC = cpu.ReadWord(ResetVector)
	cpu.SP = StackReset
	cpu.SR = StatusUnused | StatusInterrupt

	cpu.cycles = 0
	cpu.TotalCycles = 0
}

func (cpu *CPU) Clock() bool {
	if cpu.cycles > 0 {
		cpu.cycles--
		return cpu.cycles <= 0
	}

	opcode := cpu.Read(cpu.PC)

	instruction := Instructions[opcode]

	address, pageCrosed := cpu.fetchOperandAddress(instruction.AddressingMode)

	args := OperationArgs{
		instruction.AddressingMode,
		address,
	}

	cpu.cycles = instruction.Cycles

	if pageCrosed {
		cpu.cycles += instruction.AdditionalCycles
	}

	cpu.PC += uint16(instruction.Size)

	instruction.operation(cpu, args)

	cpu.TotalCycles += uint64(cpu.cycles)

	cpu.cycles--

	return false
}

func (cpu *CPU) fetchOperandAddress(addrMode AddressingMode) (uint16, bool) {
	switch addrMode {
	// Instruction's operand is implict to the intrustion or does not exist.
	case AddressingModeImplied:
		return 0, false

	// Instruction's operand is within the accumulator.
	case AddressingModeAccumulator:
		return 0, false

	// Address of instructions operand is immediately adjacent to the current program counter address.
	case AddressingModeImmediate:
		return cpu.PC + 1, false

	// Address of instructions operand is the 8-bit value in the address immediately adjacent to
	// the current program counter address mapped to the 0th page.
	case AddressingModeZeroPage:
		return uint16(cpu.Read(cpu.PC + 1)), false

	// The same as zero page address but with X registers value applied as offset
	// Mask applied to the 8 MSB in case of overflow due to offset.
	case AddressingModeZeroPageX:
		return uint16(cpu.Read(cpu.PC+1)+cpu.X) & 0x00FF, false

	// The same as zero page address but with Y registers value applied as offset
	// Mask applied to the 8 MSB in case of overflow due to offset.
	case AddressingModeZeroPageY:
		return uint16(cpu.Read(cpu.PC+1)+cpu.Y) & 0x00FF, false

	// Special addressing mode for branching operations, the 8-bit value found at the address
	// directly adjacent to the instructions operator is treated as a signed 2's compliment number
	// (-128 to 127). This offset is then applied to the address following that to jump to an address
	// within a 128 range around the address of the inital 8-bit value.
	case AddressingModeRelative:
		relAddr := uint16(cpu.Read(cpu.PC + 1))

		if relAddr&0x80 != 0 {
			relAddr |= 0xFF00
		}

		return cpu.PC + 2 + relAddr, false

	// Reads the 2 bytes starting from the address directly adjacent to the operator and treats
	// that 16-bit number as an absoulte address.
	case AddressingModeAbsolute:
		return cpu.ReadWord(cpu.PC + 1), false

	// Same as absolute but the 16-bit address has the contents of the X register applied to it
	// as an offset.
	case AddressingModeAbsoluteX:
		address := cpu.ReadWord(cpu.PC+1) + uint16(cpu.X)
		pageCrossed := cpu.pageCrossed(address-uint16(cpu.X), address)

		return address, pageCrossed

	// Same as absolute but the 16-bit address has the contents of the Y register applied to it
	// as an offset.
	case AddressingModeAbsoluteY:
		address := cpu.ReadWord(cpu.PC+1) + uint16(cpu.Y)
		pageCrossed := cpu.pageCrossed(address, address-uint16(cpu.Y))

		return address, pageCrossed

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

		return cpu.readWordbug(ptrAddr), false

	case AddressingModeIndirectX:
		ptrAddr := (uint16(cpu.Read(cpu.PC+1)) + uint16(cpu.X)) & 0x00FF

		return cpu.readWordbug(ptrAddr), false

	case AddressingModeIndirectY:
		ptrAddr := uint16(cpu.Read(cpu.PC + 1))

		address := cpu.readWordbug(ptrAddr) + uint16(cpu.Y)
		pageCrossed := cpu.pageCrossed(address, address-uint16(cpu.Y))

		return address, pageCrossed

	default:
		panic(fmt.Sprintf("Invalid addressing mode %d", addrMode))
	}
}

func (cpu *CPU) pageCrossed(a, b uint16) bool {
	return a&0xFF00 != b&0xFF00
}

func (cpu *CPU) PrintCPUState(hexidecimal bool) {
	cpu.PrintRegisters()
	cpu.PrintProcessorStatus(hexidecimal)
}

func (cpu *CPU) PrintRegisters() {
	fmt.Printf(
		"Accumulator register: 0x%02X\nIndex register X: 0x%02X\nIndex register Y: 0x%02X\nProgram counter: 0x%04X\nStack pointer: 0x%02X\n",
		cpu.A, cpu.X, cpu.Y, cpu.PC, cpu.SP,
	)
}

func (cpu *CPU) PrintProcessorStatus(hexidecimal bool) {
	if hexidecimal {
		fmt.Printf("Processor status register: 0x%02X\n", cpu.SR)
	} else {
		fmt.Printf("NVUBDIZC\n")
		fmt.Printf("%08b\n", cpu.SR)
	}
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

// Returns 16 bit value from memory similar to ReadWord, emulating the 6502 hardware bug
// where the high byte is read from the start of the low bytes page if the low byte is
// on page boudary instead of the next page i.e. the low byte of the address passed in
// is FF
func (cpu *CPU) readWordbug(addr uint16) uint16 {
	if addr&0x00FF == 0x00FF {
		return uint16(cpu.Read(addr&0xFF00))<<8 | uint16(cpu.Read(addr))
	} else {
		return cpu.ReadWord(addr)
	}
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

func (cpu *CPU) branch(branch bool, address uint16) {
	if branch {
		cpu.cycles += 1

		if cpu.pageCrossed(address, cpu.PC) {
			cpu.cycles += 1
		}

		cpu.PC = address
	}
}

// ---------- //
// Interrupts //
// ---------- //

func (cpu *CPU) IRQ() {
	if !cpu.getStatus(StatusInterrupt) {
		cpu.pushWord(cpu.PC)

		cpu.setStatus(StatusBreak, false)
		cpu.setStatus(StatusInterrupt, true)
		cpu.setStatus(StatusUnused, true)

		cpu.push(uint8(cpu.SR))

		cpu.PC = cpu.ReadWord(IRQVector)

		cpu.cycles = 7
		cpu.TotalCycles += uint64(cpu.cycles)
	}
}

func (cpu *CPU) NMI() {
	cpu.pushWord(cpu.PC)

	cpu.setStatus(StatusBreak, false)
	cpu.setStatus(StatusInterrupt, true)
	cpu.setStatus(StatusUnused, true)

	cpu.push(uint8(cpu.SR))

	cpu.PC = cpu.ReadWord(NMIVector)

	cpu.cycles = 8
	cpu.TotalCycles += uint64(cpu.cycles)
}

// --------------- //
// Offical Opcodes //
// --------------- //

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
* Logical and operand with contents of accumulator
* Set zero status if resulting value is 0
* Set negative status if resulting value's 7th bit is set
*
*/
func and(cpu *CPU, args OperationArgs) {
	cpu.A &= cpu.Read(args.address)
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
	branched := !cpu.getStatus(StatusCarry)
	cpu.branch(branched, args.address)
}

/*
*
Branch if Carry Set
* Checks if carry bit in status register is set, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func bcs(cpu *CPU, args OperationArgs) {
	branched := cpu.getStatus(StatusCarry)
	cpu.branch(branched, args.address)
}

/*
*
Branch if Equal
* Checks if zero bit in status register is set, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func beq(cpu *CPU, args OperationArgs) {
	branched := cpu.getStatus(StatusZero)
	cpu.branch(branched, args.address)
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
	branched := cpu.getStatus(StatusNegative)
	cpu.branch(branched, args.address)
}

/*
*
Branch if Not Equal
* Checks if zero bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func bne(cpu *CPU, args OperationArgs) {
	branched := !cpu.getStatus(StatusZero)
	cpu.branch(branched, args.address)
}

/*
*
Branch if Positive
* Checks if negative bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func bpl(cpu *CPU, args OperationArgs) {
	branched := !cpu.getStatus(StatusNegative)
	cpu.branch(branched, args.address)
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
	branched := !cpu.getStatus(StatusOverflow)
	cpu.branch(branched, args.address)
}

/*
*
Branch if Overflow is Set
* Checks if overflow bit in status register is clear, if so sets the program counter
register to the pre-calculated relative address in address argument.
*
*/
func bvs(cpu *CPU, args OperationArgs) {
	branched := cpu.getStatus(StatusOverflow)
	cpu.branch(branched, args.address)
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
	cpu.setStatus(StatusBreak, false)
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
	cpu.setZN(cpu.X)
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

// ----------------- //
// Unoffical Opcodes //
// ----------------- //

func ahx(cpu *CPU, args OperationArgs) {
	cpu.Write(args.address, cpu.A&cpu.X&(uint8(args.address>>8)+1))
}

/*
*
Load Accumulator and Logical Shift Right
* Performs equvilant of a immediate mode LDA
* Then performs equvilant of LSR on accumulator
* Sets Zero and Negative bits in status registers based on result
*
*/
func alr(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.Read(args.address)
	cpu.setStatus(StatusCarry, cpu.A&0x0001 != 0)
	cpu.A >>= 1
	cpu.setZN(cpu.A)
}

/*
*
AND with Accumulator and Copy N to C
* Performs the equivilant of immediate mode AND
* Sets Zero and Negative bits in status registers based on result
* Copys Negative status bit to Carry status bit
*
*/
func anc(cpu *CPU, arg OperationArgs) {
	cpu.A &= cpu.Read(arg.address)
	cpu.setZN(cpu.A)
	cpu.setStatus(StatusCarry, cpu.getStatus(StatusNegative))
}

/*
*
AND with Accumulator and Rotate Right
* Performs the equivilant of immediate mode AND
* Performs the equivilant of a ROR on the accumulator
* Sets Zero and Negative bits in status registers based on result
* Sets Carry bit in status register based on the results 6th bit
* Sets Overflow bit in status register based on the results 6th bit xor with 5th bit
*
*/
func arr(cpu *CPU, args OperationArgs) {
	cpu.A &= cpu.Read(args.address)
	cpu.A = cpu.A>>1 | cpu.A&0x0001<<7
	cpu.setZN(cpu.A)

	carryBit := cpu.A&0x20 != 0
	cpu.setStatus(StatusCarry, carryBit)

	overflowBit := ((cpu.A & 0x20 >> 5) ^ (cpu.A & 0x10 >> 4)) != 0
	cpu.setStatus(StatusOverflow, overflowBit)
}

func axs(cpu *CPU, args OperationArgs) {
	operand := uint16(cpu.Read(args.address))
	cpu.X &= cpu.A
	result := uint16(cpu.X) - operand

	carryBit := result&0xFF00 != 0
	cpu.setStatus(StatusCarry, carryBit)
	cpu.setZN(uint8(result))

	cpu.X = uint8(result)
}

func dcp(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address) - 1
	cpu.Write(args.address, operand)

	cpu.setStatus(StatusCarry, cpu.A >= operand)
	cpu.setZN(cpu.A - operand)
}

func isc(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address) + 1
	cpu.Write(args.address, operand)

	subtrahend := uint16(operand) ^ 0x00FF
	carryBit := uint16(util.Btou8(cpu.getStatus(StatusCarry)))

	result := uint16(cpu.A) + subtrahend + carryBit

	overflowed := ((uint16(cpu.A) ^ result) & ^(uint16(cpu.A) ^ subtrahend) & 0x0080) != 0

	cpu.setStatus(StatusOverflow, overflowed)
	cpu.setStatus(StatusCarry, result > 255)
	cpu.setZN(uint8(result))

	cpu.A = uint8(result)
}

func las(cpu *CPU, args OperationArgs) {
	cpu.SP &= cpu.Read(args.address)
	cpu.A = cpu.SP
	cpu.X = cpu.SP
}

func lax(cpu *CPU, args OperationArgs) {
	cpu.A = cpu.Read(args.address)
	cpu.X = cpu.A
	cpu.setZN(cpu.A)
}

func rla(cpu *CPU, args OperationArgs) {
	carryBit := util.Btou8(cpu.getStatus(StatusCarry))
	operand := cpu.Read(args.address)

	cpu.setStatus(StatusCarry, operand&0x80 != 0)
	operand = operand<<1 | carryBit
	cpu.Write(args.address, operand)

	cpu.A &= operand
	cpu.setZN(cpu.A)
}

func rra(cpu *CPU, args OperationArgs) {
	carryBit := util.Btou8(cpu.getStatus(StatusCarry))
	operand := cpu.Read(args.address)

	cpu.setStatus(StatusCarry, operand&0x01 != 0)
	operand = operand>>1 | carryBit<<7
	cpu.Write(args.address, operand)

	result := uint16(cpu.A) + uint16(operand) + uint16(util.Btou8(cpu.getStatus(StatusCarry)))

	overflowed := ((uint16(cpu.A) ^ result) & ^(uint16(cpu.A) ^ uint16(operand)) & 0x0080) != 0

	cpu.setStatus(StatusOverflow, overflowed)
	cpu.setStatus(StatusCarry, result > 255)
	cpu.setZN(uint8(result))

	cpu.A = uint8(result)
}

func sax(cpu *CPU, args OperationArgs) {
	cpu.Write(args.address, cpu.A&cpu.X)
}

func shx(cpu *CPU, args OperationArgs) {
	cpu.Write(args.address, cpu.X&(uint8(args.address>>8)+1))
}

func shy(cpu *CPU, args OperationArgs) {
	cpu.Write(args.address, cpu.Y&(uint8(args.address>>8)+1))
}

func slo(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address)

	cpu.setStatus(StatusCarry, operand&0x80 != 0)
	operand <<= 1

	cpu.Write(args.address, operand)

	cpu.A |= operand
	cpu.setZN(cpu.A)
}

func sre(cpu *CPU, args OperationArgs) {
	operand := cpu.Read(args.address)

	cpu.setStatus(StatusCarry, operand&0x01 != 0)
	operand >>= 1

	cpu.Write(args.address, operand)

	cpu.A ^= operand
	cpu.setZN(cpu.A)
}

func tas(cpu *CPU, args OperationArgs) {
	cpu.SR = Status(cpu.A & cpu.X)
	cpu.Write(args.address, uint8(cpu.SR)&(uint8(args.address>>8)+1))
}

// Unimplemented operation function
func xxx(cpu *CPU, args OperationArgs) {

}
