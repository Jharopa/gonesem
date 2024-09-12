package ppu

type (
	Ctrl   uint8
	Mask   uint8
	Status uint8
)

const (
	CtrlNametableAddressX     Ctrl = 1 << iota // N 1: Add 256 to X scroll position
	CtrlNametableAddressY                      // N 1: Add 240 to Y scroll position
	CtrlIncrementMode                          // I 0: Add 1 across; 1: Add 32 down
	CtrlSpriteTableAddress                     // S 0: Address $0000; 1: Address $1000; ignore in 8x16
	CtrlBackgroundTableAddres                  // B 0: Address $0000; 1: Address $1000
	CtrlSpriteSize                             // H 0: 8x8 pixles; 1: 8x16 pixels
	CtrlMasterSlaveMode                        // P 0: read backfrop from EXT pins; 1: output color on EXT pins
	CtrlGenerateNMI                            // V 0: off; 1: on
)

const (
	MaskGreyscale          Mask = 1 << iota // g
	MaskShowBackgroundLeft                  // m Show background in the leftmost 8 pixles of screen
	MaskShowSpritesLeft                     // M Show sprites in the leftmost 8 pixles of screen
	MaskShowBackground                      // b
	MaskShowSprites                         // s
	MaskEmphasizeRed                        // R
	MaskEmphasizeGreen                      // G
	MaskEmphasizeBlue                       // B
)

const (
	StatusOpenBus       Status = 1 << 5 // O
	StatusSpriteZeroHit Status = 1 << 6 // S
	StatusVerticalBlank Status = 1 << 7 // V
)

func (ppu *PPU) setCtrl(ctrl Ctrl, value bool) {
	if value {
		ppu.ctrl |= ctrl
	} else {
		ppu.ctrl &^= ctrl
	}
}

func (ppu *PPU) getCtrl(ctrl Ctrl) bool {
	return (ppu.ctrl & ctrl) != 0
}

func (ppu *PPU) setMask(mask Mask, value bool) {
	if value {
		ppu.mask |= mask
	} else {
		ppu.mask &^= mask
	}
}

func (ppu *PPU) getMask(mask Mask) bool {
	return (ppu.mask & mask) != 0
}

func (ppu *PPU) setStatus(status Status, value bool) {
	if value {
		ppu.status |= status
	} else {
		ppu.status &^= status
	}
}

func (ppu *PPU) getStatus(status Status) bool {
	return (ppu.status & status) != 0
}
