package cartridge

import "fmt"

type Mapper interface {
	PGRRead(addr uint16, memory []uint8) uint8
	PRGWrite(addr uint16, value uint8, memory []uint8)
	CHRRead(addr uint16, memory []uint8) uint8
	CHRWrite(addr uint16, value uint8, memory []uint8)
}

func NewMapper(mapperID uint8) Mapper {
	switch mapperID {
	case 0:
		return Mapper000{}
	default:
		panic(fmt.Sprintf("Unsupported mapper, ID %d", mapperID))
	}
}
