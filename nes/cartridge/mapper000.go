package cartridge

type Mapper000 struct {
	cartridge *Cartridge
}

func (mapper Mapper000) PGRRead(addr uint16) uint8 {
	if addr >= 0x8000 && addr <= 0xFFFF {
		pgrMemorySize := len(mapper.cartridge.pgrMemory)
		return mapper.cartridge.pgrMemory[addr%uint16(pgrMemorySize)]
	}

	return 0
}

// Mapper000 PRG rom only, no writing
func (mapper Mapper000) PRGWrite(addr uint16, value uint8) {
}

func (mapper Mapper000) CHRRead(addr uint16) uint8 {
	if addr <= 0x1FFF {
		return mapper.cartridge.chrMemory[addr]
	}

	return 0
}

// Mapper000 CHR rom only, no writing
func (mapper Mapper000) CHRWrite(addr uint16, value uint8) {
}
