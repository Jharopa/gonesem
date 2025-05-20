package memory

type MirrorMode uint8

const (
	MirorHorizontal MirrorMode = iota
	MirrorVertical
)

var MirroredNametableIdxLookup = [...][4]uint16{
	{0, 0, 1, 1},
	{0, 1, 0, 1},
}

/*
*
Takes a MirrorMode supplied by the iNES header and read/write address in the PPU
and returns which nametable index that address should mirror to.
See: https://www.nesdev.org/wiki/Mirroring#Nametable_Mirroring
*/
func MirroredNametableIdx(mode MirrorMode, addr uint16) uint16 {
	tableIdx := (addr - 0x2000) / 0x400

	return MirroredNametableIdxLookup[mode][tableIdx]
}
