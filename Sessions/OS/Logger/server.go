package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Logger safely writes concurrent log messages to a file
type Logger struct {
	file     *os.File
	logChan  chan string
	wg       sync.WaitGroup
	stopOnce sync.Once
	stopChan chan struct{}
}

// NewLogger initializes the logger with a target log file path
func NewLogger(filePath string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		file:     file,
		logChan:  make(chan string, 1000),
		stopChan: make(chan struct{}),
	}

	logger.wg.Add(1)
	go logger.startWriter()
	return logger, nil
}

// startWriter listens on the channel and writes logs to disk
func (l *Logger) startWriter() {
	defer l.wg.Done()
	for {
		select {
		case logMsg := <-l.logChan:
			if !endsWithNewline(logMsg) {
				logMsg += "\n"
			}
			_, _ = l.file.WriteString(logMsg) // Ignore write error for brevity
		case <-l.stopChan:
			// Drain channel before exiting
			for {
				select {
				case logMsg := <-l.logChan:
					if !endsWithNewline(logMsg) {
						logMsg += "\n"
					}
					_, _ = l.file.WriteString(logMsg)
				default:
					return
				}
			}
		}
	}
}

// Log queues a log message for writing
func (l *Logger) Log(appID string, message string) {
	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, appID, message)
	l.logChan <- logLine
}

// Close shuts down the logger and flushes all logs
func (l *Logger) Close() {
	l.stopOnce.Do(func() {
		close(l.stopChan)
		l.wg.Wait()
		l.file.Close()
	})
}

// endsWithNewline checks if the string ends with \n
func endsWithNewline(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '\n'
}

// main simulates multiple apps writing logs concurrently
func main() {
	logger, err := NewLogger("central.log")
	if err != nil {
		fmt.Println("❌ Error initializing logger:", err)
		return
	}
	defer logger.Close()

	apps := []string{"auth-service", "payment-service", "frontend", "worker", "analytics"}
	var wg sync.WaitGroup

	// Simulate each app sending logs concurrently
	for _, app := range apps {
		wg.Add(1)
		go func(appID string) {
			defer wg.Done()
			for i := 1; i <= 10; i++ {
				logger.Log(appID, fmt.Sprintf("Log entry %d", i))
				time.Sleep(time.Duration(100+10*i) * time.Millisecond)
			}
		}(app)
	}

	wg.Wait()
	fmt.Println("✅ All logs submitted. Check 'central.log'.")
}
