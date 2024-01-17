package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "example")
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return
	}
	defer tmpFile.Close()

	// Write some data to the temporary file
	data := []byte("Hello, temporary file!")
	_, err = tmpFile.Write(data)
	if err != nil {
		fmt.Println("Error writing to temporary file:", err)
		return
	}

	// Seek to the beginning of the file before reading
	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Println("Error seeking to the beginning of the file:", err)
		return
	}

	// Open the temporary file for reading
	readFile, err := os.Open(tmpFile.Name())
	if err != nil {
		fmt.Println("Error opening temporary file for reading:", err)
		return
	}
	defer readFile.Close()

	// Read from the file
	readData := make([]byte, len(data))
	_, err = readFile.Read(readData)
	if err != nil {
		fmt.Println("Error reading from temporary file:", err)
		return
	}

	// Print the read data
	fmt.Println("Read data:", string(readData))
}
