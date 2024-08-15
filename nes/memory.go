package nes

type Memory interface {
	Read(addr uint16) uint8
	Write(addr uint16, value uint8)
}