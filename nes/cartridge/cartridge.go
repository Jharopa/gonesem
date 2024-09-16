package cartridge

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// iNES header
type Header struct {
	NESConst   [4]uint8 // Constant $4E $45 $53 $1A (ASCII "NES" followed by MS-DOS EOF)
	PRGSize    uint8    // Size of PRG ROM in 16kb units
	CHRSize    uint8    // Size of CHR ROM in 8kb units
	Mapper1    uint8    // Lower nibble of mapper ID
	Mapper2    uint8    // Higher nibble of mapper ID
	PRGRamSize uint8    // PRG RAM size
	TvSystem1  uint8    // TV System type
	TvSystem2  uint8    // TV System type
	_          [5]uint8 // Unused in iNES 1.0 format
}

type Cartridge struct {
	pgrBanks   uint8
	chrBanks   uint8
	mirrorMode uint8
	pgrMemory  []uint8
	chrMemory  []uint8
	mapper     Mapper
}

func NewCartridge(romPath string) (*Cartridge, error) {
	romFile, err := os.Open(romPath)

	if err != nil {
		return nil, fmt.Errorf("failed to open ROM file: %s", err)
	}

	defer romFile.Close()

	header := Header{}

	if err := binary.Read(romFile, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read in header from rom file: %s", err)
	}

	cartridge := &Cartridge{}

	mapperID := (header.Mapper1 & 0xF0) | header.Mapper2>>4
	hasTrainer := header.Mapper1>>2&0x01 != 0

	cartridge.mapper = NewMapper(mapperID, cartridge)

	if hasTrainer {
		if _, err = romFile.Seek(512, io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to skip trainer data: %s", err)
		}
	}

	cartridge.pgrBanks = header.PRGSize
	cartridge.chrBanks = header.CHRSize

	cartridge.pgrMemory = make([]uint8, uint32(cartridge.pgrBanks)*16384)

	if _, err := io.ReadFull(romFile, cartridge.pgrMemory); err != nil {
		return nil, fmt.Errorf("failed to read PRG data into PRG ROM memory: %s", err)
	}

	cartridge.chrMemory = make([]uint8, uint32(cartridge.chrBanks)*8192)

	if _, err := io.ReadFull(romFile, cartridge.chrMemory); err != nil {
		return nil, fmt.Errorf("failed to read CHR data into CHR ROM memory: %s", err)
	}

	return cartridge, nil
}

func (cartridge *Cartridge) PRGRead(addr uint16) uint8 {
	return cartridge.mapper.PGRRead(addr)
}

func (cartridge *Cartridge) PRGWrite(addr uint16, value uint8) {
	cartridge.mapper.PRGWrite(addr, value)
}

func (cartridge *Cartridge) CHRRead(addr uint16) uint8 {
	return cartridge.mapper.CHRRead(addr)
}

func (cartridge *Cartridge) CHRWrite(addr uint16, value uint8) {
	cartridge.mapper.CHRWrite(addr, value)
}
