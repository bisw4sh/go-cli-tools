package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	input := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	text, _ := input.ReadByte()
	fmt.Println("You entered:", text)
}
