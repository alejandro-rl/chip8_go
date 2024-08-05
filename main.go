package main

import (
	"fmt"
)

// b byte
// Create a mask with a single bit set at the current position
//mask := byte(1 << 5)
// Check if the bit at position i is set
//bit := (b & mask) >> 7
//Set bit to 1
//b = b | mask
//Set bit to 0
//b = b & (^mask)

func PrintDisplay(display [32][64]int) {
	for _, j := range display {
		fmt.Print(j, "\t")
		fmt.Println()
	}
	fmt.Println()

}

func main() {

	chip8 := NewChip()
	chip8.LoadROM("./roms/IBM Logo.ch8")

	// //fmt.Printf("%d\n", chip8.program_counter)

	for {
		//time.Sleep(time.Second)
		chip8.Cycle()
		PrintDisplay(chip8.display)

	}

}
