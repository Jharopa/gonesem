package cartridge

type Mapper000 struct {
}

func (mapper Mapper000) PGRRead(addr uint16, memory []uint8) uint8 {
	if addr >= 0x8000 && addr <= 0xFFFF {
		return memory[addr%uint16(len(memory))]
	}

	return 0
}

// Mapper000 PRG rom only, no writing
func (mapper Mapper000) PRGWrite(addr uint16, value uint8, memory []uint8) {
}

func (mapper Mapper000) CHRRead(addr uint16, memory []uint8) uint8 {
	if addr <= 0x1FFF {
		return memory[addr]
	}

	return 0
}

// Mapper000 CHR rom only, no writing
func (mapper Mapper000) CHRWrite(addr uint16, value uint8, memory []uint8) {
}
