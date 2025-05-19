package ppu

func (ppu *PPU) getNametableByte() {
	ppu.nameTableByte = ppu.read(0x2000 | (uint16(ppu.vramAddr) & 0x0FFF))
}

func (ppu *PPU) getAttributeTableByte() {
	ppu.attributeTableByte = ppu.read(0x23C0 | (ppu.vramAddr & 0x0C00) | ((ppu.vramAddr >> 0x04) & 0x38) | ((ppu.vramAddr >> 0x02) & 0x07))

	if (ppu.vramAddr>>5)&0x02 == 0x02 {
		ppu.attributeTableByte >>= 4
	}

	if ppu.vramAddr&0x02 == 0x02 {
		ppu.attributeTableByte >>= 2
	}

	ppu.attributeTableByte &= 0x03
}

func (ppu *PPU) getPatternTableByte(low bool) uint8 {
	var (
		tableIdx    uint16
		fineY       uint16
		planeOffset uint16
		address     uint16
	)

	if !ppu.getCtrl(CtrlBackgroundTableAddress) {
		tableIdx = 0
	} else {
		tableIdx = 1
	}

	fineY = (ppu.vramAddr >> 12) & 0x07

	if low {
		planeOffset = 0
	} else {
		planeOffset = 8
	}

	address = tableIdx*0x1000 + uint16(ppu.nameTableByte)*16 + fineY + planeOffset

	return ppu.read(address)
}

func (ppu *PPU) getPatternTableByteLow() {
	ppu.patternTableByteLow = ppu.getPatternTableByte(true)
}

func (ppu *PPU) getPatternTableByteHigh() {
	ppu.patternTableByteHigh = ppu.getPatternTableByte(false)
}
