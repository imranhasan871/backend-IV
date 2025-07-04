package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Customer struct {
	ID    int
	Email string
}

type EmailResult struct {
	CustomerID int
	Email      string
	Success    bool
	Error      error
}

// Simulated sendEmail function (I/O-bound task)
func sendEmail(c Customer) error {
	// Simulate network I/O delay
	time.Sleep(time.Duration(100+rand.Intn(300)) * time.Millisecond)

	// Simulate random failure
	if rand.Float32() < 0.1 {
		return fmt.Errorf("failed to send to %s", c.Email)
	}
	return nil
}

// Worker to send email jobs
func emailWorker(id int, jobs <-chan Customer, results chan<- EmailResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for customer := range jobs {
		log.Printf("[Worker %d] Sending email to %s", id, customer.Email)
		err := sendEmail(customer)
		results <- EmailResult{
			CustomerID: customer.ID,
			Email:      customer.Email,
			Success:    err == nil,
			Error:      err,
		}
	}
}

// Simulate fetching all customers who ordered today
func fetchTodaysCustomers() []Customer {
	customers := make([]Customer, 1000)
	for i := range customers {
		customers[i] = Customer{
			ID:    i + 1,
			Email: fmt.Sprintf("user%03d@example.com", i+1),
		}
	}
	return customers
}

// The scheduled job that runs monthly
func sendMonthlyInvoices() {
	log.Println("📤 Starting monthly invoice email job")

	customers := fetchTodaysCustomers()

	const workerCount = 20
	jobs := make(chan Customer, workerCount*2)
	results := make(chan EmailResult, len(customers))

	var wg sync.WaitGroup

	// Start workers
	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go emailWorker(i, jobs, results, &wg)
	}

	// Send jobs
	go func() {
		for _, c := range customers {
			jobs <- c
		}
		close(jobs)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and report results
	success := 0
	failed := 0

	for result := range results {
		if result.Success {
			log.Printf("✅ Sent to %s", result.Email)
			success++
		} else {
			log.Printf("❌ Failed to send to %s: %v", result.Email, result.Error)
			failed++
		}
	}

	log.Printf("🎉 Done. Total: %d, Success: %d, Failed: %d\n", len(customers), success, failed)
}

func main() {
	c := cron.New()

	// Schedule: At 2:00 AM on the 1st of every month
	c.AddFunc("0 2 1 * *", sendMonthlyInvoices)

	log.Println("📅 Invoice email scheduler started")
	c.Start()

	// ➕ Trigger one run immediately
	go sendMonthlyInvoices()

	select {} // keep the program running
}
