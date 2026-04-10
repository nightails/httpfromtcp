package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buff := make([]byte, 8)

	for {
		n, err := io.ReadFull(file, buff)
		if n > 0 {
			fmt.Printf("read: %s\n", string(buff[:n]))
		}

		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			panic(err)
		}
	}
}
