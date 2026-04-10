package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buff := make([]byte, 8)
	var currentLine string

	for {
		n, err := io.ReadFull(file, buff)
		if n > 0 {
			b := string(buff[:n])
			if strings.Contains(b, "\n") {
				parts := strings.Split(b, "\n")
				for i := 0; i < len(parts)-1; i++ {
					fmt.Printf("read: %s\n", currentLine+parts[i])
				}
				currentLine = parts[len(parts)-1]
			} else {
				currentLine += b
			}
		}

		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			panic(err)
		}
	}

	if currentLine != "" {
		fmt.Printf("read: %s\n", currentLine)
	}
}
