package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

const (
	batchSize  = 700000
	numWorkers = 10
)

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
	fileNamePrefix := fileName[:len(fileName)-4]

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	outputChannels := make([]chan string, numWorkers)
	for i := 0; i < numWorkers; i++ {
		outputChannels[i] = make(chan string)
	}

	writerPool := sync.Pool{
		New: func() interface{} {
			file, err := os.Create(fmt.Sprintf("%s_%d.txt", fileNamePrefix, fileCount))
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				return nil
			}
			return bufio.NewWriter(file)
		},
	}

	filePool := sync.Pool{
		New: func() interface{} {
			file, err := os.Create(fmt.Sprintf("%s_%d.txt", fileNamePrefix, fileCount))
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				return nil
			}
			return file
		},
	}

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			writer := writerPool.Get().(*bufio.Writer)
			defer writerPool.Put(writer)

			file := filePool.Get().(*os.File)
			defer filePool.Put(file)

			lineCount := 0

			for str := range outputChannels[workerID] {
				if lineCount == batchSize {
					writer.Flush()
					fileCount++
					file.Close()

					file = filePool.Get().(*os.File)
					defer filePool.Put(file)

					writer = writerPool.Get().(*bufio.Writer)
					defer writerPool.Put(writer)

					lineCount = 0
				}

				_, err := writer.WriteString(str + "\n")
				if err != nil {
					fmt.Printf("Error writing to file: %v\n", err)
					return
				}

				lineCount++
			}

			writer.Flush()
		}(i)
	}

	generate("", allowedChars, func(str string) {
		workerID := hashString(str) % numWorkers
		outputChannels[workerID] <- str
	})

	for i := 0; i < numWorkers; i++ {
		close(outputChannels[i])
	}

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

func hashString(str string) int {
	hash := 0
	for _, char := range str {
		hash = (hash*31 + int(char)) % numWorkers
	}
	return hash
}
