package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
)

const batchSize = 700000

func main() {
	var allowedChars string

	fmt.Print("Enter the allowed characters: ")
	fmt.Scan(&allowedChars)

	fileName := "output.txt"
	err := saveStringsToSeparateFiles(fileName, allowedChars)
	if err != nil {
		fmt.Printf("Error saving strings to files: %v\n", err)
		return
	}

	fmt.Printf("Strings saved to separate files.\n")
}

func saveStringsToSeparateFiles(fileName string, allowedChars string) error {
	fileCount := 0
	fileNamePrefix := fileName[:len(fileName)-4] // Remove the ".txt" extension

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		file, err := os.Create(fileNamePrefix + "_" + strconv.Itoa(fileCount) + ".txt")
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			return
		}

		writer := bufio.NewWriter(file)
		lineCount := 0

		generate("", allowedChars, func(str string) {
			if lineCount == batchSize {
				writer.Flush()
				fileCount++
				file.Close()

				file, err = os.Create(fileNamePrefix + "_" + strconv.Itoa(fileCount) + ".txt")
				if err != nil {
					fmt.Printf("Error creating file: %v\n", err)
					return
				}

				writer = bufio.NewWriter(file)
				lineCount = 0
			}

			_, err := writer.WriteString(str + "\n")
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				return
			}

			lineCount++
		})

		writer.Flush()
		file.Close()
	}()

	wg.Wait()
	return nil
}

func generate(prefix, allowedChars string, processFunc func(string)) {
	if len(allowedChars) == 0 {
		processFunc(prefix)
		return
	}

	for i, char := range allowedChars {
		newAllowedChars := allowedChars[:i] + allowedChars[i+1:]
		newPrefix := prefix + string(char)
		generate(newPrefix, newAllowedChars, processFunc)
	}
}
