package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

/*
- Result hold the outcome of a single file hashing operation.
*/
type Result struct {
	Path     string
	Hash     string
	Err      error
	FileType string
	Depth    int
}

/*
WORKER FUCNTION
- fucntion will be called by goroutines.
- function will read file paths from the 'jobs' channel, calculates the hash,
and sends a Result to the 'results' channel.
*/
func worker(jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	// Decrement the WaitGroup counter when the goroutuine finishes.
	defer wg.Done()

	// Loop over the jobs chgannel until it's closed and empty,
	for path := range jobs {
		file, err := os.Open(path)
		if err != nil {
			results <- Result{Path: path, Err: err}
			continue
		}

		// Create new SHA-256 hasher.
		hasher := sha256.New()

		// Copy file content into hasher
		// io.copy is efficient for large files.
		if _, err := io.Copy(hasher, file); err != nil {
			file.Close() // Ensure file is closed on error
			results <- Result{Path: path, Err: err}
			continue
		}

		file.Close()

		// Get the resulting hash and format uit as a hax string
		hash := fmt.Sprintf("%x", hasher.Sum(nil))
		results <- Result{Path: path, Hash: hash}
	}
}

func main() {
	// 1- Get directory path from command line
	if len(os.Args) < 2 {
		log.Fatal("Please provide a dir path.")
	}
	rootDier := os.Args[1]

	// Use the number of CPU cores for number of woeker goroutine.
	numWorkers := runtime.NumCPU()

	jobs := make(chan string, numWorkers)
	results := make(chan Result, 100) // buffer 100 for result channel
	var wg sync.WaitGroup
	// 2- Start the worker goroutines.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// Structured Loggin
	logFile, err := os.OpenFile("file-hashes.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open lop file:: %v", err)
	}

	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	// 3- Walk the directory and sen file paths to the jobs channel,
	// run this in separate goroutine so that the main goroutine can
	// proccess the resules as the come in

	go func() {
		// this filepath.Walk is a blocking call.
		filepath.Walk(rootDier, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// we only want to hash regular files, not dir.
			if !info.IsDir() {
				jobs <- path
			}
			return nil
		})

		close(jobs)
	}()

	// 4- Wait for all workers to finish, then close the results channel
	// We need another goroutine to do this so we don't block the results collection loop.
	go func() {
		wg.Wait()
		close(results)
	}()

	// 5- Collect and print all reuslts.
	fmt.Println("File processing complete. See file-hashing.log for derails.")
	for res := range results {
		res.FileType = filepath.Ext(res.Path)
		res.Depth = len(strings.Split(res.Path, string(os.PathSeparator))) - 1

		if res.Err != nil {
			logger.Error("Failed to Hash the file",
				"path", res.Path,
				"error", res.Err.Error(),
			)
		} else {
			logger.Info("Successfully hashed file",
				"path", res.Path,
				"sha256", res.Hash,
				"file-type", res.FileType,
				"depth", res.Depth,
			)
		}
	}
}
