package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const (
	url       = "http://127.0.0.1:8081/largefile.bin" // Replace with real file URL
	partCount = 4                                     // Number of parallel downloads
	output    = "downloaded_file.zip"
)

func main() {
	// Step 1: Get file size
	size, err := getFileSize(url)
	if err != nil {
		fmt.Println("❌ Failed to get file size:", err)
		return
	}
	fmt.Printf("📦 File size: %d bytes\n", size)

	// Step 2: Create output file of the correct size
	file, err := os.Create(output)
	if err != nil {
		fmt.Println("❌ Error creating file:", err)
		return
	}
	defer file.Close()
	err = file.Truncate(size)
	if err != nil {
		fmt.Println("❌ Error resizing file:", err)
		return
	}

	// Step 3: Download in parts
	var wg sync.WaitGroup
	chunkSize := size / int64(partCount)

	for i := 0; i < partCount; i++ {
		wg.Add(1)
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == partCount-1 {
			end = size - 1 // Last chunk goes to end
		}

		go func(start, end int64, partNum int) {
			defer wg.Done()
			err := downloadPart(url, file, start, end, partNum)
			if err != nil {
				fmt.Printf("❌ Part %d failed: %v\n", partNum, err)
			} else {
				fmt.Printf("✅ Part %d downloaded [%d - %d]\n", partNum, start, end)
			}
		}(start, end, i)
	}

	wg.Wait()
	fmt.Println("🎉 Download completed successfully!")
}

// getFileSize sends a HEAD request to get Content-Length
func getFileSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned %s", resp.Status)
	}

	sizeStr := resp.Header.Get("Content-Length")
	if sizeStr == "" {
		return 0, fmt.Errorf("no Content-Length header")
	}
	return strconv.ParseInt(sizeStr, 10, 64)
}

// downloadPart downloads a specific byte range and writes to the file at the right position
func downloadPart(url string, file *os.File, start, end int64, partNum int) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
	req.Header.Set("Range", rangeHeader)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	// Write to specific offset
	buffer := make([]byte, 32*1024) // 32 KB buffer
	offset := start
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.WriteAt(buffer[:n], offset)
			if writeErr != nil {
				return writeErr
			}
			offset += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}
