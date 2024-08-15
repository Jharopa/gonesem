package nes

import "fmt"

type Status uint8

const (
	StatusCarry Status = 1 << iota 	// C
	StatusZero 						// Z
	StatusInterrupt 				// I
	StatusDecimal 					// D
	StatusBreak 					// B
	StatusUnused 					// U
	StatusOverflow 					// V
	StatusNegative 					// N
)

const (
	StackPage uint16 = 0x0100
	StackReset uint8 = 0xFD
)

type AddressingMode uint8

const (
	AddressingModeImplied AddressingMode = iota // IMP 0
	AddressingModeAccumulator					// ACC 1
	AddressingModeImmediate 					// IMM 2
	AddressingModeZeroPage 						// ZP0 3
	AddressingModeZeroPageX 					// ZPX 4
	AddressingModeZeroPageY 					// ZPY 5
	AddressingModeRelative 						// REL 6
	AddressingModeAbsolute 						// ABS 7
	AddressingModeAbsoluteX 					// ABX 8
	AddressingModeAbsoluteY						// ABY 9
	AddressingModeIndirect 						// IND 10
	AddressingModeIndirectX 					// IZX 11
	AddressingModeIndirectY 					// IZY 12
)

var mnemonics = [256] string {
//   0      1      2      3      4      5      6      7      8      9      A      B      C      D      E      F
	"BRK", "ORA", "???", "???", "???", "ORA", "ASL", "???", "PHP", "ORA", "ASL", "???", "???", "ORA", "ASL", "???", // 0
	"BPL", "ORA", "???", "???", "???", "ORA", "ASL", "???", "CLC", "ORA", "???", "???", "???", "ORA", "ASL", "???", // 1
	"JSR", "AND", "???", "???", "BIT", "AND", "ROL", "???", "PLP", "AND", "ROL", "???", "BIT", "AND", "ROL", "???", // 2
	"BMI", "AND", "???", "???", "???", "AND", "ROL", "???", "SEC", "AND", "???", "???", "???", "AND", "ROL", "???", // 3
	"RTI", "EOR", "???", "???", "???", "EOR", "LSR", "???", "PHA", "EOR", "LSR", "???", "JMP", "EOR", "LSR", "???", // 4
	"BVC", "EOR", "???", "???", "???", "EOR", "LSR", "???", "CLI", "EOR", "???", "???", "???", "EOR", "LSR", "???", // 5
	"RTS", "ADC", "???", "???", "???", "ADC", "ROR", "???", "PLA", "ADC", "ROR", "???", "JMP", "ADC", "ROR", "???", // 6
	"BVS", "ADC", "???", "???", "???", "ADC", "ROR", "???", "SEI", "ADC", "???", "???", "???", "ADC", "ROR", "???", // 7
	"???", "STA", "???", "???", "STY", "STA", "STX", "???", "DEY", "???", "TXA", "???", "STY", "STA", "STX", "???", // 8
	"BCC", "STA", "???", "???", "STY", "STA", "STX", "???", "TYA", "STA", "TXS", "???", "???", "STA", "???", "???", // 9
	"LDY", "LDA", "LDX", "???", "LDY", "LDA", "LDX", "???", "TAY", "LDA", "TAX", "???", "LDY", "LDA", "LDX", "???", // A
	"BCS", "LDA", "???", "???", "LDY", "LDA", "LDX", "???", "CLV", "LDA", "TSX", "???", "LDY", "LDA", "LDX", "???", // B
	"CPY", "CMP", "???", "???", "CPY", "CMP", "DEC", "???", "INY", "CMP", "DEX", "???", "CPY", "CMP", "DEC", "???", // C
	"BNE", "CMP", "???", "???", "???", "CMP", "DEC", "???", "CLD", "CMP", "???", "???", "???", "CMP", "DEC", "???", // D
	"CPX", "SBC", "???", "???", "CPX", "SBC", "INC", "???", "INX", "SBC", "NOP", "???", "CPX", "SBC", "INC", "???", // E
	"BEQ", "SBC", "???", "???", "???", "SBC", "INC", "???", "SED", "SBC", "???", "???", "???", "SBC", "INC", "???", // F
}


type CPU struct {
	A uint8 	// Accumulator register
	X uint8 	// X index register
	Y uint8 	// Y index register
	PC uint16 	// Program counter register
	SP uint8 	// Statck pointer register
	SR Status 	// Status register

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
	cpu.PC = cpu.ReadWord(0xFFCC)
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
	lo := uint8(value & 0X00FF)
	cpu.Write(addr, lo)
	cpu.Write(addr + 1, hi)
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
	cpu.Write(StackPage + uint16(cpu.SP), value)
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
	return hi << 8 | lo
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
	return cpu.SR & status != 0
}

func (cpu *CPU) SetZ(value uint8) {
	if value == 0 {
		cpu.SetStatus(StatusZero, true)
	} else {
		cpu.SetStatus(StatusZero, false)
	}
}

func (cpu *CPU) SetN(value uint8) {
	if value & 0x80 != 0 {
		cpu.SetStatus(StatusNegative, true)
	} else {
		cpu.SetStatus(StatusNegative, false)
	}
}

func (cpu *CPU) SetZN(value uint8) {
	cpu.SetZ(value)
	cpu.SetN(value)
}