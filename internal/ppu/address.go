package ppu

/*
Increments current vram address along the scanline. If background or foreground
rendering are enabled, checks coarse x value in address is equal to 31 resetting
coarse x to 0 and flipping the x nametablebit;
Otherwise increments coarse x by 1.
*/
func (ppu *PPU) incrementScrollX() {
	if ppu.vramAddr&0x1F == 0x1F {
		ppu.vramAddr &= 0xFFE0
		ppu.vramAddr ^= 0x0400
	} else {
		ppu.vramAddr++
	}
}

/*
Increments current vram address down the scanline. If background or foreground
rendering are enabled, increments fine y by 1 if it is less than 7;
Otherwise accounting for the nametables attribute memory boundary,
either resets coarse y and/or flips the y nametable bit, or increments coarse y by 1.
*/
func (ppu *PPU) incrementScrollY() {
	if ppu.vramAddr&0x7000 != 0x7000 {
		ppu.vramAddr += 0x1000
	} else {
		ppu.vramAddr &= 0x8FFF

		y := (ppu.vramAddr & 0x03E0) >> 5

		if y == 0x1D { // Coarse y is 29
			y = 0
			ppu.vramAddr ^= 0x0800
		} else if y == 0x1F { // Coarse y is 31
			y = 0
		} else {
			y++
		}

		ppu.vramAddr = (ppu.vramAddr & 0xFC1F) | (y << 5)
	}
}

func (ppu *PPU) copyAddressX() {
	ppu.vramAddr = (ppu.vramAddr & 0x7BE0) | (ppu.tramAddr & 0x041F)
}

func (ppu *PPU) copyAddressY() {
	ppu.vramAddr = (ppu.vramAddr & 0x041F) | (ppu.tramAddr & 0x7BE0)
}
