package cpu

import "gonesem/internal/util"

// --------------- //
// Offical Opcodes //
// --------------- //

/*
Add with carry
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
