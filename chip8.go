package main

import (
	"fmt"
	"os"
)

type Chip8 struct {

	// Registers - 16 1-byte registers called V0 to VF
	registers [16]byte

	//Program Counter (PC) - points to current instruction in memory
	program_counter uint16

	//Index Register (I) - points to locations in memory
	index_register uint16

	// Stack - to call and return from subroutines
	stack [16]uint16

	// Delay timer -  is decremented at a rate of 60 Hz (60 times per second) until it reaches 0
	delay_timer uint8

	// Sound timer - functions like the delay timer, but which also gives off a beeping sound as long as it’s not 0
	sound_timer uint8

	// Memory - 4kB of RAM
	// CHIP-8’s index register and program counter can only address 12 bits
	memory [4096]byte

	//Display - 64 x 32 pixels, monochromatic
	display [32][64]int

	//Keypad -  16 keys
	keypad [16]uint16
}

func NewChip() *Chip8 {
	chip := new(Chip8)

	// Fontset - to represent sprites
	fontset := [80]byte{
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

	chip.program_counter = 0x200

	// Load Fontset

	for i := 0; i < 80; i++ {
		chip.memory[i] = fontset[i]
	}

	return chip

}

// LoadROM receives a path to a ROM, tries to load it into memory and returns true if it was sucessful
func (chip *Chip8) LoadROM(path string) bool {

	// Read contents of file

	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Print("Could not read file \n")
		return false
	}

	// Load in memory from 0x200(512) onwards.
	mem_value := chip.program_counter

	//First, check if the ROM is too big to load.
	if (int(mem_value) + len(data)) >= len(chip.memory) {

		fmt.Print("File size too big to fit into memory! \n")
		return false
	}

	//If it's not, load it into memory.
	for _, byte := range data {
		chip.memory[mem_value] = byte
		mem_value++

	}

	return true

}

// Executes one cycle.

func (chip *Chip8) Cycle() {

	// The opcode has 2 bytes, but our memory has 1 byte values, to address this:
	//		First, add 8 zeroes to the right of the byte in memory where the program counter points to.
	//		Then, make a bitwise_or operation to add the next byte in memory to those zeroes.

	opcode := int(uint16(chip.memory[chip.program_counter])<<8 | uint16(chip.memory[chip.program_counter+1]))

	//Get first nibble of opcode
	opcode_nibble_1 := GetNibbles(opcode, 12, 0xF000)

	var val, reg1, reg2 int

	//Instruction Set
	switch opcode_nibble_1 {

	//0x00E0 - Clear the display.
	case 0:
		chip.display = [len(chip.display)][len(chip.display[0])]int{}
		chip.program_counter += 2

	//1NNN - Jump to location NNN
	case 1:
		chip.program_counter = uint16(GetNibbles(opcode, 0, 0x0FFF))

	//6XNN - Set V[X] = NN
	case 6:
		//Get value to set (NN)
		val = GetNibbles(opcode, 0, 0x00FF)
		//Get register index
		reg1 = GetNibbles(opcode, 8, 0x0F00)

		chip.registers[reg1] = byte(val)
		chip.program_counter += 2

	//7XNN - Set V[X] = V[X] + NN
	case 7:
		//Get value to set (NN)
		val = GetNibbles(opcode, 0, 0x00FF)
		//Get register index
		reg1 = GetNibbles(opcode, 8, 0x0F00)

		chip.registers[reg1] += byte(val)
		chip.program_counter += 2

	// ANNN - Set Index Register  I = NNN
	case 10:
		//Get Value to set (NNN)
		val = GetNibbles(opcode, 0, 0x0FFF)
		chip.index_register = uint16(val)
		chip.program_counter += 2

	//DXYN - Display n-byte sprite starting at memory location I at (V[X], V[Y]), set V[F] = collision.
	case 13:

		//get X and Y coordinates from the registers
		reg1 = GetNibbles(opcode, 8, 0x0F00)
		reg2 = GetNibbles(opcode, 4, 0x00F0)

		x := chip.registers[reg1]
		y := chip.registers[reg2]

		//Get the number of bytes
		n_bytes := GetNibbles(opcode, 0, 0x000F)

		// The starting position of the sprite will wrap around the screen.

		x = x & 63
		y = y & 31

		//V[F] should be set to zero.
		chip.registers[15] = 0

		for i := range n_bytes {

			// Get the Nth byte of the sprite
			// counting from memory address the Index Register.
			sprite_byte := chip.memory[chip.index_register+uint16(i)]

			// Iterate over every bit, from left to right.
			for j := 7; j >= 0; j-- {

				// Create a mask with a single bit set at the current position.
				mask := byte(1 << j)
				// Check if the bit at position i is set.
				bit := (sprite_byte & mask) >> j

				//If the current bit is on and the pixel in x,y is also on, turn it off
				//and set V[F] = 1

				if bit == 1 && chip.display[y][x] == 1 {
					sprite_byte = sprite_byte & (^mask)
					chip.registers[15] = 1

					// If the current pixel in the sprite row is on and the screen pixel is not, draw the pixel
					// at the X and Y coordinates.
				} else {
					chip.display[y][x] = 1

				}
				// If you reach the right edge of the screen, stop drawing this row.
				if x > 62 {
					break
				}

				// Increment X
				x++
			}

			//Increment Y
			y++

			//Stop if you reach the bottom edge of the screen.
			if y > 30 {
				break
			}

		}

		chip.program_counter += 2

	default:
		fmt.Print("Invalid Opcode\n")

	}

}

//Extract nibbles from opcode.

func GetNibbles(val int, bits int, binary_and int) int {

	return ((val & binary_and) >> bits)

}
