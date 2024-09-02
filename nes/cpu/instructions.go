package cpu

type Instruction struct {
	operation                   func(*CPU, OperationArgs)
	Mnemonic                    string
	AddressingMode              AddressingMode
	InstructionSize             uint8
	InstructionCycles           uint8
	AdditionalInstructionCycles uint8
}

var Instructions = [256]Instruction{
	{brk, "BRK", AddressingModeImplied, 1, 7, 0},     // 0x00
	{ora, "ORA", AddressingModeIndirectX, 2, 6, 0},   // 0x01
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x02
	{slo, "SLO", AddressingModeIndirectX, 2, 8, 0},   // 0x03
	{nop, "NOP", AddressingModeZeroPage, 2, 3, 0},    // 0x04
	{ora, "ORA", AddressingModeZeroPage, 2, 3, 0},    // 0x05
	{asl, "ASL", AddressingModeZeroPage, 2, 5, 0},    // 0x06
	{slo, "SLO", AddressingModeZeroPage, 2, 5, 0},    // 0x07
	{php, "PHP", AddressingModeImplied, 1, 3, 0},     // 0x08
	{ora, "ORA", AddressingModeImmediate, 2, 2, 0},   // 0x09
	{asl, "ASL", AddressingModeAccumulator, 1, 2, 0}, // 0x0A
	{anc, "ANC", AddressingModeImmediate, 2, 2, 0},   // 0x0B
	{nop, "NOP", AddressingModeAbsolute, 3, 4, 0},    // 0x0C
	{ora, "ORA", AddressingModeAbsolute, 3, 4, 0},    // 0x0D
	{asl, "ASL", AddressingModeAbsolute, 3, 6, 0},    // 0x0E
	{slo, "SLO", AddressingModeAbsolute, 3, 6, 0},    // 0x0F
	{bpl, "BPL", AddressingModeRelative, 2, 2, 0},    // 0x10
	{ora, "ORA", AddressingModeIndirectY, 2, 5, 1},   // 0x11
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x12
	{slo, "SLO", AddressingModeIndirectY, 2, 8, 0},   // 0x13
	{nop, "NOP", AddressingModeZeroPageX, 2, 4, 0},   // 0x14
	{ora, "ORA", AddressingModeZeroPageX, 2, 4, 0},   // 0x15
	{asl, "ASL", AddressingModeZeroPageX, 2, 6, 0},   // 0x16
	{slo, "SLO", AddressingModeZeroPageX, 2, 6, 0},   // 0x17
	{clc, "CLC", AddressingModeImplied, 1, 2, 0},     // 0x18
	{ora, "ORA", AddressingModeAbsoluteY, 3, 4, 1},   // 0x19
	{nop, "NOP", AddressingModeImplied, 1, 2, 0},     // 0x1A
	{slo, "SLO", AddressingModeAbsoluteY, 3, 7, 0},   // 0x1B
	{nop, "NOP", AddressingModeAbsoluteX, 3, 4, 1},   // 0x1C
	{ora, "ORA", AddressingModeAbsoluteX, 3, 4, 1},   // 0x1D
	{asl, "ASL", AddressingModeAbsoluteX, 3, 7, 0},   // 0x1E
	{slo, "SLO", AddressingModeAbsoluteX, 3, 7, 0},   // 0x1F
	{jsr, "JSR", AddressingModeAbsolute, 3, 6, 0},    // 0x20
	{and, "AND", AddressingModeIndirectX, 2, 6, 0},   // 0x21
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x22
	{rla, "RLA", AddressingModeIndirectX, 2, 8, 0},   // 0x23
	{bit, "BIT", AddressingModeZeroPage, 2, 3, 0},    // 0x24
	{and, "AND", AddressingModeZeroPage, 2, 3, 0},    // 0x25
	{rol, "ROL", AddressingModeZeroPage, 2, 5, 0},    // 0x26
	{rla, "RLA", AddressingModeZeroPage, 2, 5, 0},    // 0x27
	{plp, "PLP", AddressingModeImplied, 1, 4, 0},     // 0x28
	{and, "AND", AddressingModeImmediate, 2, 2, 0},   // 0x29
	{rol, "ROL", AddressingModeAccumulator, 1, 2, 0}, // 0x2A
	{anc, "ANC", AddressingModeImmediate, 2, 2, 0},   // 0x2B
	{bit, "BIT", AddressingModeAbsolute, 3, 4, 0},    // 0x2C
	{and, "AND", AddressingModeAbsolute, 3, 4, 0},    // 0x2D
	{rol, "ROL", AddressingModeAbsolute, 3, 6, 0},    // 0x2E
	{rla, "RLA", AddressingModeAbsolute, 3, 6, 0},    // 0x2F
	{bmi, "BMI", AddressingModeRelative, 2, 2, 0},    // 0x30
	{and, "AND", AddressingModeIndirectY, 2, 5, 1},   // 0x31
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x32
	{rla, "RLA", AddressingModeIndirectY, 2, 8, 0},   // 0x33
	{nop, "NOP", AddressingModeZeroPageX, 2, 4, 0},   // 0x34
	{and, "AND", AddressingModeZeroPageX, 2, 4, 0},   // 0x35
	{rol, "ROL", AddressingModeZeroPageX, 2, 6, 0},   // 0x36
	{rla, "RLA", AddressingModeZeroPageX, 2, 6, 0},   // 0x37
	{sec, "SEC", AddressingModeImplied, 1, 2, 0},     // 0x38
	{and, "AND", AddressingModeAbsoluteY, 3, 4, 1},   // 0x39
	{nop, "NOP", AddressingModeImplied, 1, 2, 0},     // 0x3A
	{rla, "RLA", AddressingModeAbsoluteY, 3, 7, 0},   // 0x3B
	{nop, "NOP", AddressingModeAbsoluteX, 3, 4, 1},   // 0x3C
	{and, "AND", AddressingModeAbsoluteX, 3, 4, 1},   // 0x3D
	{rol, "ROL", AddressingModeAbsoluteX, 3, 7, 0},   // 0x3E
	{rla, "RLA", AddressingModeAbsoluteX, 3, 7, 0},   // 0x3F
	{rti, "RTI", AddressingModeImplied, 1, 6, 0},     // 0x40
	{eor, "EOR", AddressingModeIndirectX, 2, 6, 0},   // 0x41
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x42
	{sre, "SRE", AddressingModeIndirectX, 2, 8, 0},   // 0x43
	{nop, "NOP", AddressingModeZeroPage, 2, 3, 0},    // 0x44
	{eor, "EOR", AddressingModeZeroPage, 2, 3, 0},    // 0x45
	{lsr, "LSR", AddressingModeZeroPage, 2, 5, 0},    // 0x46
	{sre, "SRE", AddressingModeZeroPage, 2, 5, 0},    // 0x47
	{pha, "PHA", AddressingModeImplied, 1, 3, 0},     // 0x48
	{eor, "EOR", AddressingModeImmediate, 2, 2, 0},   // 0x49
	{lsr, "LSR", AddressingModeAccumulator, 1, 2, 0}, // 0x4A
	{alr, "ALR", AddressingModeImmediate, 2, 2, 0},   // 0x4B
	{jmp, "JMP", AddressingModeAbsolute, 3, 3, 0},    // 0x4C
	{eor, "EOR", AddressingModeAbsolute, 3, 4, 0},    // 0x4D
	{lsr, "LSR", AddressingModeAbsolute, 3, 6, 0},    // 0x4E
	{sre, "SRE", AddressingModeAbsolute, 3, 6, 0},    // 0x4F
	{bvc, "BVC", AddressingModeRelative, 2, 2, 0},    // 0x50
	{eor, "EOR", AddressingModeIndirectY, 2, 5, 1},   // 0x51
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x52
	{sre, "SRE", AddressingModeIndirectY, 2, 8, 0},   // 0x53
	{nop, "NOP", AddressingModeZeroPageX, 2, 4, 0},   // 0x54
	{eor, "EOR", AddressingModeZeroPageX, 2, 4, 0},   // 0x55
	{lsr, "LSR", AddressingModeZeroPageX, 2, 6, 0},   // 0x56
	{sre, "SRE", AddressingModeZeroPageX, 2, 6, 0},   // 0x57
	{cli, "CLI", AddressingModeImplied, 1, 2, 0},     // 0x58
	{eor, "EOR", AddressingModeAbsoluteY, 3, 4, 1},   // 0x59
	{nop, "NOP", AddressingModeImplied, 1, 2, 0},     // 0x5A
	{sre, "SRE", AddressingModeAbsoluteY, 3, 7, 0},   // 0x5B
	{nop, "NOP", AddressingModeAbsoluteX, 3, 4, 1},   // 0x5C
	{eor, "EOR", AddressingModeAbsoluteX, 3, 4, 1},   // 0x5D
	{lsr, "LSR", AddressingModeAbsoluteX, 3, 7, 0},   // 0x5E
	{sre, "SRE", AddressingModeAbsoluteX, 3, 7, 0},   // 0x5F
	{rts, "RTS", AddressingModeImplied, 1, 6, 0},     // 0x60
	{adc, "ADC", AddressingModeIndirectX, 2, 6, 0},   // 0x61
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x62
	{rra, "RRA", AddressingModeIndirectX, 2, 8, 0},   // 0x63
	{nop, "NOP", AddressingModeZeroPage, 2, 3, 0},    // 0x64
	{adc, "ADC", AddressingModeZeroPage, 2, 3, 0},    // 0x65
	{ror, "ROR", AddressingModeZeroPage, 2, 5, 0},    // 0x66
	{rra, "RRA", AddressingModeZeroPage, 2, 5, 0},    // 0x67
	{pla, "PLA", AddressingModeImplied, 1, 4, 0},     // 0x68
	{adc, "ADC", AddressingModeImmediate, 2, 2, 0},   // 0x69
	{ror, "ROR", AddressingModeAccumulator, 1, 2, 0}, // 0x6A
	{arr, "ARR", AddressingModeImmediate, 2, 2, 0},   // 0x6B
	{jmp, "JMP", AddressingModeIndirect, 3, 5, 0},    // 0x6C
	{adc, "ADC", AddressingModeAbsolute, 3, 4, 0},    // 0x6D
	{ror, "ROR", AddressingModeAbsolute, 3, 6, 0},    // 0x6E
	{rra, "RRA", AddressingModeAbsolute, 3, 6, 0},    // 0x6F
	{bvs, "BVS", AddressingModeRelative, 2, 2, 0},    // 0x70
	{adc, "ADC", AddressingModeIndirectY, 2, 5, 1},   // 0x71
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x72
	{rra, "RRA", AddressingModeIndirectY, 2, 8, 0},   // 0x73
	{nop, "NOP", AddressingModeZeroPageX, 2, 4, 0},   // 0x74
	{adc, "ADC", AddressingModeZeroPageX, 2, 4, 0},   // 0x75
	{ror, "ROR", AddressingModeZeroPageX, 2, 6, 0},   // 0x76
	{rra, "RRA", AddressingModeZeroPageX, 2, 6, 0},   // 0x77
	{sei, "SEI", AddressingModeImplied, 1, 2, 0},     // 0x78
	{adc, "ADC", AddressingModeAbsoluteY, 3, 4, 1},   // 0x79
	{nop, "NOP", AddressingModeImplied, 1, 2, 0},     // 0x7A
	{rra, "RRA", AddressingModeAbsoluteY, 3, 7, 0},   // 0x7B
	{nop, "NOP", AddressingModeAbsoluteX, 3, 4, 1},   // 0x7C
	{adc, "ADC", AddressingModeAbsoluteX, 3, 4, 1},   // 0x7D
	{ror, "ROR", AddressingModeAbsoluteX, 3, 7, 0},   // 0x7E
	{rra, "RRA", AddressingModeAbsoluteX, 3, 7, 0},   // 0x7F
	{nop, "NOP", AddressingModeImmediate, 2, 2, 0},   // 0x80
	{sta, "STA", AddressingModeIndirectX, 2, 6, 0},   // 0x81
	{nop, "NOP", AddressingModeImmediate, 2, 2, 0},   // 0x82
	{sax, "SAX", AddressingModeIndirectX, 2, 6, 0},   // 0x83
	{sty, "STY", AddressingModeZeroPage, 2, 3, 0},    // 0x84
	{sta, "STA", AddressingModeZeroPage, 2, 3, 0},    // 0x85
	{stx, "STX", AddressingModeZeroPage, 2, 3, 0},    // 0x86
	{sax, "SAX", AddressingModeZeroPage, 2, 3, 0},    // 0x87
	{dey, "DEY", AddressingModeImplied, 1, 2, 0},     // 0x88
	{nop, "NOP", AddressingModeImmediate, 2, 2, 0},   // 0x89
	{txa, "TXA", AddressingModeImplied, 1, 2, 0},     // 0x8A
	{xxx, "XAA", AddressingModeImplied, 0, 0, 0},     // 0x8B Unimplemented
	{sty, "STY", AddressingModeAbsolute, 3, 4, 0},    // 0x8C
	{sta, "STA", AddressingModeAbsolute, 3, 4, 0},    // 0x8D
	{stx, "STX", AddressingModeAbsolute, 3, 4, 0},    // 0x8E
	{sax, "SAX", AddressingModeAbsolute, 3, 4, 0},    // 0x8F
	{bcc, "BCC", AddressingModeRelative, 2, 2, 0},    // 0x90
	{sta, "STA", AddressingModeIndirectY, 2, 6, 0},   // 0x91
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0x92
	{ahx, "AHX", AddressingModeIndirectY, 2, 6, 0},   // 0x93
	{sty, "STY", AddressingModeZeroPageX, 2, 4, 0},   // 0x94
	{sta, "STA", AddressingModeZeroPageX, 2, 4, 0},   // 0x95
	{stx, "STX", AddressingModeZeroPageY, 2, 4, 0},   // 0x96
	{sax, "SAX", AddressingModeZeroPageY, 2, 4, 0},   // 0x97
	{tya, "TYA", AddressingModeImplied, 1, 2, 0},     // 0x98
	{sta, "STA", AddressingModeAbsoluteY, 3, 5, 0},   // 0x99
	{txs, "TXS", AddressingModeImplied, 1, 2, 0},     // 0x9A
	{tas, "TAS", AddressingModeAbsoluteY, 3, 5, 0},   // 0x9B
	{shy, "SHY", AddressingModeAbsoluteX, 3, 5, 0},   // 0x9C
	{sta, "STA", AddressingModeAbsoluteX, 3, 5, 0},   // 0x9D
	{shx, "SHX", AddressingModeAbsoluteY, 3, 5, 0},   // 0x9E
	{ahx, "AHX", AddressingModeAbsoluteY, 3, 5, 0},   // 0x9F
	{ldy, "LDY", AddressingModeImmediate, 2, 2, 0},   // 0xA0
	{lda, "LDA", AddressingModeIndirectX, 2, 6, 0},   // 0xA1
	{ldx, "LDX", AddressingModeImmediate, 2, 2, 0},   // 0xA2
	{lax, "LAX", AddressingModeIndirectX, 2, 6, 0},   // 0xA3
	{ldy, "LDY", AddressingModeZeroPage, 2, 3, 0},    // 0xA4
	{lda, "LDA", AddressingModeZeroPage, 2, 3, 0},    // 0xA5
	{ldx, "LDX", AddressingModeZeroPage, 2, 3, 0},    // 0xA6
	{lax, "LAX", AddressingModeZeroPage, 2, 3, 0},    // 0xA7
	{tay, "TAY", AddressingModeImplied, 1, 2, 0},     // 0xA8
	{lda, "LDA", AddressingModeImmediate, 2, 2, 0},   // 0xA9
	{tax, "TAX", AddressingModeImplied, 1, 2, 0},     // 0xAA
	{xxx, "LXA", AddressingModeImmediate, 2, 2, 0},   // 0xAB Unimplemented
	{ldy, "LDY", AddressingModeAbsolute, 3, 4, 0},    // 0xAC
	{lda, "LDA", AddressingModeAbsolute, 3, 4, 0},    // 0xAD
	{ldx, "LDX", AddressingModeAbsolute, 3, 4, 0},    // 0xAE
	{lax, "LAX", AddressingModeAbsolute, 3, 4, 0},    // 0xAF
	{bcs, "BCS", AddressingModeRelative, 2, 2, 0},    // 0xB0
	{lda, "LDA", AddressingModeIndirectY, 2, 5, 1},   // 0xB1
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0xB2
	{lax, "LAX", AddressingModeIndirectY, 2, 5, 1},   // 0xB3
	{ldy, "LDY", AddressingModeZeroPageX, 2, 4, 0},   // 0xB4
	{lda, "LDA", AddressingModeZeroPageX, 2, 4, 0},   // 0xB5
	{ldx, "LDX", AddressingModeZeroPageY, 2, 4, 0},   // 0xB6
	{lax, "LAX", AddressingModeZeroPageY, 2, 4, 0},   // 0xB7
	{clv, "CLV", AddressingModeImplied, 1, 2, 0},     // 0xB8
	{lda, "LDA", AddressingModeAbsoluteY, 3, 4, 1},   // 0xB9
	{tsx, "TSX", AddressingModeImplied, 1, 2, 0},     // 0xBA
	{las, "LAS", AddressingModeAbsoluteY, 3, 4, 1},   // 0xBB
	{ldy, "LDY", AddressingModeAbsoluteX, 3, 4, 1},   // 0xBC
	{lda, "LDA", AddressingModeAbsoluteX, 3, 4, 1},   // 0xBD
	{ldx, "LDX", AddressingModeAbsoluteY, 3, 4, 1},   // 0xBE
	{lax, "LAX", AddressingModeAbsoluteY, 3, 4, 1},   // 0xBF
	{cpy, "CPY", AddressingModeImmediate, 2, 2, 0},   // 0xC0
	{cmp, "CMP", AddressingModeIndirectX, 2, 6, 0},   // 0xC1
	{nop, "NOP", AddressingModeImmediate, 2, 2, 0},   // 0xC2
	{dcp, "DCP", AddressingModeIndirectX, 2, 8, 1},   // 0xC3
	{cpy, "CPY", AddressingModeZeroPage, 2, 3, 0},    // 0xC4
	{cmp, "CMP", AddressingModeZeroPage, 2, 3, 0},    // 0xC5
	{dec, "DEC", AddressingModeZeroPage, 2, 5, 0},    // 0xC6
	{dcp, "DCP", AddressingModeZeroPage, 2, 5, 0},    // 0xC7
	{iny, "INY", AddressingModeImplied, 1, 2, 0},     // 0xC8
	{cmp, "CMP", AddressingModeImmediate, 2, 2, 0},   // 0xC9
	{dex, "DEX", AddressingModeImplied, 1, 2, 0},     // 0xCA
	{axs, "AXS", AddressingModeImmediate, 2, 2, 0},   // 0xCB
	{cpy, "CPY", AddressingModeAbsolute, 3, 4, 0},    // 0xCC
	{cmp, "CMP", AddressingModeAbsolute, 3, 4, 0},    // 0xCD
	{dec, "DEC", AddressingModeAbsolute, 3, 6, 0},    // 0xCE
	{dcp, "DCP", AddressingModeAbsolute, 3, 6, 0},    // 0xCF
	{bne, "BNE", AddressingModeRelative, 2, 2, 0},    // 0xD0
	{cmp, "CMP", AddressingModeIndirectY, 2, 5, 1},   // 0xD1
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0xD2
	{dcp, "DCP", AddressingModeIndirectY, 2, 8, 0},   // 0xD3
	{nop, "NOP", AddressingModeZeroPageX, 2, 4, 0},   // 0xD4
	{cmp, "CMP", AddressingModeZeroPageX, 2, 4, 0},   // 0xD5
	{dec, "DEC", AddressingModeZeroPageX, 2, 6, 0},   // 0xD6
	{dcp, "DCP", AddressingModeZeroPageX, 2, 6, 0},   // 0xD7
	{cld, "CLD", AddressingModeImplied, 1, 2, 0},     // 0xD8
	{cmp, "CMP", AddressingModeAbsoluteY, 3, 4, 1},   // 0xD9
	{nop, "NOP", AddressingModeImplied, 1, 2, 0},     // 0xDA
	{dcp, "DCP", AddressingModeAbsoluteY, 3, 7, 0},   // 0xDB
	{nop, "NOP", AddressingModeAbsoluteX, 3, 4, 1},   // 0xDC
	{cmp, "CMP", AddressingModeAbsoluteX, 3, 4, 1},   // 0xDD
	{dec, "DEC", AddressingModeAbsoluteX, 3, 7, 0},   // 0xDE
	{dcp, "DCP", AddressingModeAbsoluteX, 3, 7, 0},   // 0xDF
	{cpx, "CPX", AddressingModeImmediate, 2, 2, 0},   // 0xE0
	{sbc, "SBC", AddressingModeIndirectX, 2, 6, 0},   // 0xE1
	{nop, "NOP", AddressingModeImmediate, 2, 2, 0},   // 0xE2
	{isc, "ISC", AddressingModeIndirectX, 2, 8, 0},   // 0xE3
	{cpx, "CPX", AddressingModeZeroPage, 2, 3, 0},    // 0xE4
	{sbc, "SBC", AddressingModeZeroPage, 2, 3, 0},    // 0xE5
	{inc, "INC", AddressingModeZeroPage, 2, 5, 0},    // 0xE6
	{isc, "ISC", AddressingModeZeroPage, 2, 5, 0},    // 0xE7
	{inx, "INX", AddressingModeImplied, 1, 2, 0},     // 0xE8
	{sbc, "SBC", AddressingModeImmediate, 2, 2, 0},   // 0xE9
	{nop, "NOP", AddressingModeImplied, 1, 2, 0},     // 0xEA
	{sbc, "SBC", AddressingModeImmediate, 2, 2, 0},   // 0xEB
	{cpx, "CPX", AddressingModeAbsolute, 3, 4, 0},    // 0xEC
	{sbc, "SBC", AddressingModeAbsolute, 3, 4, 0},    // 0xED
	{inc, "INC", AddressingModeAbsolute, 3, 6, 0},    // 0xEE
	{isc, "ISC", AddressingModeAbsolute, 3, 6, 0},    // 0xEF
	{beq, "BEQ", AddressingModeRelative, 2, 2, 0},    // 0xF0
	{sbc, "SBC", AddressingModeIndirectY, 2, 5, 1},   // 0xF1
	{nop, "STP", AddressingModeImplied, 1, 0, 0},     // 0xF2
	{isc, "ISC", AddressingModeIndirectY, 2, 8, 0},   // 0xF3
	{nop, "NOP", AddressingModeZeroPageX, 2, 4, 0},   // 0xF4
	{sbc, "SBC", AddressingModeZeroPageX, 2, 4, 0},   // 0xF5
	{inc, "INC", AddressingModeZeroPageX, 2, 6, 0},   // 0xF6
	{isc, "ISC", AddressingModeZeroPageX, 2, 6, 0},   // 0xF7
	{sed, "SED", AddressingModeImplied, 1, 2, 0},     // 0xF8
	{sbc, "SBC", AddressingModeAbsoluteY, 3, 4, 1},   // 0xF9
	{nop, "NOP", AddressingModeImplied, 1, 2, 0},     // 0xFA
	{isc, "ISC", AddressingModeAbsoluteY, 3, 7, 0},   // 0xFB
	{nop, "NOP", AddressingModeAbsoluteX, 3, 4, 1},   // 0xFC
	{sbc, "SBC", AddressingModeAbsoluteX, 3, 4, 1},   // 0xFD
	{inc, "INC", AddressingModeAbsoluteX, 3, 7, 0},   // 0xFE
	{isc, "ISC", AddressingModeAbsoluteX, 3, 7, 0},   // 0xFF
}
