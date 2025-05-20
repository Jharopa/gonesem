package cartridge

type Mapper000 struct {
	cartridge *Cartridge
}

func (mapper Mapper000) PGRRead(addr uint16) uint8 {
	if addr >= 0x8000 {
		var pgrMemorySize uint16

		if mapper.cartridge.pgrBanks > 1 {
			pgrMemorySize = 0x7FFF
		} else {
			pgrMemorySize = 0x3FFF
		}

		return mapper.cartridge.pgrMemory[addr&pgrMemorySize]
	}

	return 0
}

func (mapper Mapper000) PRGWrite(addr uint16, value uint8) {
	if addr >= 0x8000 {
		var pgrMemorySize uint16

		if mapper.cartridge.pgrBanks > 1 {
			pgrMemorySize = 0x7FFF
		} else {
			pgrMemorySize = 0x3FFF
		}

		mapper.cartridge.pgrMemory[addr&pgrMemorySize] = value
	}
}

func (mapper Mapper000) CHRRead(addr uint16) uint8 {
	if addr <= 0x1FFF {
		return mapper.cartridge.chrMemory[addr]
	}

	return 0
}

func (mapper Mapper000) CHRWrite(addr uint16, value uint8) {
	if addr <= 0x1FFF {
		if mapper.cartridge.chrBanks == 0 {
			mapper.cartridge.chrMemory[addr] = value
		}
	}
}
