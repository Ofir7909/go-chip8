package main

import (
	"log"
	"math/rand/v2"
)

const screenWidth uint8 = 64
const screenHeight uint8 = 32
const vramSize uint16 = uint16(screenWidth) * uint16(screenHeight)
const programBaseAddress uint16 = 0x200

var fontset = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Cpu struct {
	memory   [4096]byte
	graphics [vramSize]uint8

	registers      [16]uint8
	programCounter uint16
	index          uint16
	delayTimer     uint8
	soundTimer     uint8

	stack        [16]uint16
	stackPointer uint16

	keys [16]uint8
}

func (cpu *Cpu) reset() {
	cpu.programCounter = programBaseAddress
	cpu.stackPointer = 0
	cpu.delayTimer = 0
	cpu.soundTimer = 0

	/*for i, v := range fontset {
		cpu.memory[i] = v
	}*/
	copy(cpu.memory[:len(fontset)], fontset[:])
}

func (cpu *Cpu) step() {
	operation := uint16(cpu.memory[cpu.programCounter])<<8 | uint16(cpu.memory[cpu.programCounter+1])
	log.Printf("%x\n", operation)

	firstNibble := operation >> 12 & 0xF
	incrementPC := true

	switch firstNibble {
	case 0x0:
		if operation == 0x00E0 {
			//CLS
			for i := range cpu.graphics {
				cpu.graphics[i] = 0x0
			}

		} else if operation == 0x00EE {
			//RET
			cpu.stackPointer -= 1
			cpu.programCounter = cpu.stack[cpu.stackPointer]
			cpu.stack[cpu.stackPointer] = 0 // optional

		} else {
			// SYS addr - Ignored
		}
	case 0x1:
		// JP addr
		cpu.programCounter = operation & 0x0FFF
		incrementPC = false

	case 0x2:
		// CALL addr
		cpu.stack[cpu.stackPointer] = cpu.programCounter
		cpu.stackPointer += 1
		cpu.programCounter = operation & 0x0FFF
		incrementPC = false

	case 0x3:
		// SE Vx, byte
		x := operation & 0x0F00 >> 8
		k := uint8(operation & 0x00FF)
		if cpu.registers[x] == k {
			cpu.programCounter += 2
		}

	case 0x4:
		// SNE Vx, byte
		x := operation & 0x0F00 >> 8
		k := uint8(operation & 0x00FF)
		if cpu.registers[x] != k {
			cpu.programCounter += 2
		}

	case 0x5:
		// SE Vx, Vy
		x := operation & 0x0F00 >> 8
		y := operation & 0x00F0 >> 4
		if cpu.registers[x] == cpu.registers[y] {
			cpu.programCounter += 2
		}

	case 0x6:
		// LD Vx, byte
		x := operation & 0x0F00 >> 8
		k := uint8(operation & 0x00FF)
		cpu.registers[x] = k

	case 0x7:
		// ADD Vx, byte
		x := operation & 0x0F00 >> 8
		k := uint8(operation & 0x00FF)
		cpu.registers[x] += k

	case 0x8:
		x := operation & 0x0F00 >> 8
		y := operation & 0x00F0 >> 4
		switch operation & 0x000F {
		case 0x0:
			// LD Vx, Vy
			cpu.registers[x] = cpu.registers[y]

		case 0x1:
			// OR Vx, Vy
			cpu.registers[x] |= cpu.registers[y]

		case 0x2:
			// AND Vx, Vy
			cpu.registers[x] &= cpu.registers[y]

		case 0x3:
			// XOR Vx, Vy
			cpu.registers[x] ^= cpu.registers[y]

		case 0x4:
			// ADD Vx, Vy
			temp := uint16(cpu.registers[x]) + uint16(cpu.registers[y])
			if temp > 255 {
				cpu.registers[0xF] = 1
			} else {
				cpu.registers[0xF] = 0
			}
			cpu.registers[x] = uint8(temp)

		case 0x5:
			// SUB Vx, Vy
			if cpu.registers[x] > cpu.registers[y] {
				cpu.registers[0xF] = 1
			} else {
				cpu.registers[0xF] = 0
			}
			cpu.registers[x] -= cpu.registers[y]

		case 0x6:
			// SHR Vx {, Vy}
			if cpu.registers[x]&0b0000_0001 == 1 {
				cpu.registers[0xF] = 1
			} else {
				cpu.registers[0xF] = 0
			}
			cpu.registers[x] >>= 1
		case 0x7:
			// SUBN Vx, Vy
			if cpu.registers[y] > cpu.registers[x] {
				cpu.registers[0xF] = 1
			} else {
				cpu.registers[0xF] = 0
			}
			cpu.registers[x] = cpu.registers[y] - cpu.registers[x]
		case 0xE:
			// SHL Vx {, Vy}
			if cpu.registers[x]&0b1000_0000 != 0 {
				cpu.registers[0xF] = 1
			} else {
				cpu.registers[0xF] = 0
			}
			cpu.registers[x] <<= 1
		}
	case 0x9:
		// SNE Vx, Vy
		x := operation & 0x0F00 >> 8
		y := operation & 0x00F0 >> 4
		if cpu.registers[x] != cpu.registers[y] {
			cpu.programCounter += 2
		}

	case 0xA:
		// LD I, addr
		cpu.index = operation & 0x0FFF

	case 0xB:
		// JP V0, addr
		n := operation & 0x0FFF
		cpu.programCounter = uint16(cpu.registers[0]) + n

	case 0xC:
		// RND Vx, byte
		x := operation & 0x0F00 >> 8
		k := uint8(operation & 0x00FF)
		r := uint8(rand.IntN(256))
		cpu.registers[x] = r & k

	case 0xD:
		// DRW Vx, Vy, nibble
		x := operation & 0x0F00 >> 8
		y := operation & 0x00F0 >> 4
		n := uint8(operation & 0x000F)

		posX := cpu.registers[x]
		posY := cpu.registers[y]

		cpu.registers[0xF] = 0
		for r := uint8(0); r < n; r++ {
			rowPixels := cpu.memory[cpu.index+uint16(r)]
			for p := uint8(0); p < 8; p++ {
				if rowPixels&(0x80>>p) != 0 {
					wrappedX := (posX + p) % screenWidth
					wrappedY := (posY + r) % screenHeight

					pixelOffset := uint16(wrappedY)*uint16(screenWidth) + uint16(wrappedX)
					if cpu.graphics[pixelOffset] != 0 {
						cpu.registers[0xF] = 1
					}
					cpu.graphics[pixelOffset] ^= 1

				}

			}
		}
	case 0xE:
		switch operation & 0x00FF {
		case 0x9E:
			// SKP Vx
			x := operation & 0x0F00 >> 8
			regX := cpu.registers[x]
			if cpu.keys[regX] == 1 {
				cpu.programCounter += 2
			}
		case 0xA1:
			// SKNP Vx
			x := operation & 0x0F00 >> 8
			if cpu.keys[cpu.registers[x]] != 1 {
				cpu.programCounter += 2
			}
		}
	case 0xF:
		x := operation & 0x0F00 >> 8

		switch operation & 0x00FF {
		case 0x07:
			// LD Vx, DT
			cpu.registers[x] = cpu.delayTimer

		case 0x0A:
			// LD Vx, K
			keyPressed := false
			for i, v := range cpu.keys {
				if v != 0 {
					keyPressed = true
					cpu.registers[x] = uint8(i)
					break
				}
			}
			if !keyPressed {
				incrementPC = false
			}

		case 0x15:
			// LD DT, Vx
			cpu.delayTimer = cpu.registers[x]

		case 0x18:
			// LD ST, Vx
			cpu.soundTimer = cpu.registers[x]

		case 0x1E:
			// ADD I, Vx
			cpu.index += uint16(cpu.registers[x])

		case 0x29:
			// LD F, Vx
			cpu.index = uint16(cpu.registers[x]) * 5

		case 0x33:
			// LD B, Vx
			valX := cpu.registers[x]
			cpu.memory[cpu.index] = valX / 100
			cpu.memory[cpu.index+1] = (valX % 100) / 10
			cpu.memory[cpu.index+2] = (valX % 10)
		case 0x55:
			// LD [I], Vx
			for i, v := range cpu.registers[:x+1] {
				cpu.memory[cpu.index+uint16(i)] = v
			}
		case 0x65:
			// LD Vx, [I]
			for i := range x + 1 {
				cpu.registers[i] = cpu.memory[cpu.index+uint16(i)]
			}
		}
	}

	if incrementPC {
		cpu.programCounter += 2
	}

	if cpu.delayTimer > 0 {
		cpu.delayTimer -= 1
	}
	if cpu.soundTimer > 0 {
		cpu.soundTimer -= 1
	}
}
