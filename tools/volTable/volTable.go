package main

import "fmt"

var step float32 = 100.0 / 127.0

func main() {
	for i := 0; i <= 127; i++ {
		fmt.Printf("\"%6.2f \", ", float32(i)*step)
	}
}
