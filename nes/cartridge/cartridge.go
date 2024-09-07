package cartridge

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
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

func newHeader(data []uint8) Header {
	header := Header{}

	err := binary.Read(bytes.NewReader(data[:]), binary.BigEndian, &header)

	if err != nil {
		log.Fatalf("Failed to load ROM header data into header type: %s", err)
	}

	return header
}

type Cartridge struct {
	pgrBanks  uint8
	chrBanks  uint8
	pgrMemory []uint8
	chrMemory []uint8
	mapper    Mapper
}

func NewCartridge(romPath string) *Cartridge {
	cartridge := &Cartridge{}

	rom := loadROM(romPath)

	header := newHeader(rom[0x00:0x10])

	mapperID := (header.Mapper1 & 0xF0) | header.Mapper2>>4
	hasTraining := header.Mapper1>>2&0x01 != 0

	cartridge.mapper = NewMapper(mapperID, cartridge)

	var trainingOffset uint16

	if hasTraining {
		trainingOffset = 512
	} else {
		trainingOffset = 0
	}

	cartridge.pgrBanks = header.PRGSize
	cartridge.chrBanks = header.CHRSize

	pgrOffset := 0x10 + trainingOffset
	cartridge.pgrMemory = rom[pgrOffset : uint16(header.PRGSize)*16384]

	chrOffset := pgrOffset + uint16(header.PRGSize)*16384
	cartridge.chrMemory = rom[chrOffset : uint16(header.CHRSize)*8192]

	return cartridge
}

func loadROM(romPath string) []uint8 {
	file, err := os.Open(romPath)

	if err != nil {
		log.Printf("Failed to open ROM file: %s", err)
	}

	stat, err := file.Stat()

	if err != nil {
		log.Printf("Failed to retrieve ROM file stats: %s", err)
	}

	rom := make([]byte, stat.Size())

	_, err = bufio.NewReader(file).Read(rom)

	if err != nil && err != io.EOF {
		log.Printf("Failed to read file into ROM buffer: %s", err)
	}

	return rom
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
