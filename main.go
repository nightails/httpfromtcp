package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const fileInput = "messages.txt"

func main() {
	file, err := os.Open(fileInput)
	if err != nil {
		log.Fatalf("error opening %s: %s\n", fileInput, err.Error())
	}
	defer file.Close()

	ch := getLineChannel(file)
	for line := range ch {
		fmt.Printf("read: %s\n", line)
	}
}

func getLineChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		defer f.Close()
		defer close(ch)

		currentLine := ""
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if currentLine != "" {
					ch <- currentLine
					currentLine = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				ch <- currentLine + parts[i]
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()
	return ch
}
