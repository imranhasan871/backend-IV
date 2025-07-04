package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type ImageJob struct {
	FileName string
}

type Result struct {
	Job ImageJob
	Err error
}

func resize(image, size string) error {
	time.Sleep(300 * time.Millisecond)
	if rand.Float32() < 0.05 {
		return errors.New("resize failed")
	}
	fmt.Printf("Resized %s to %s size\n", image, size)
	return nil
}

func watermark(image string) error {
	time.Sleep(200 * time.Millisecond)
	if rand.Float32() < 0.05 {
		return errors.New("watermark failed")
	}
	fmt.Printf("Watermarked %s\n", image)
	return nil
}

func saveImage(image string) error {
	time.Sleep(500 * time.Millisecond)
	if rand.Float32() < 0.1 {
		return errors.New("save failed")
	}
	fmt.Printf("Saved %s to storage\n", image)
	return nil
}

func processImage(job ImageJob) error {
	sizes := []string{"thumbnail", "medium", "large"}
	for _, size := range sizes {
		if err := resize(job.FileName, size); err != nil {
			return fmt.Errorf("resize %s failed: %w", size, err)
		}
	}
	if err := watermark(job.FileName); err != nil {
		return fmt.Errorf("watermark failed: %w", err)
	}
	return nil
}

func processorWorker(id int, jobs <-chan ImageJob, saveQueue chan<- ImageJob, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		log.Printf("[Processor %d] Processing %s", id, job.FileName)
		if err := processImage(job); err != nil {
			results <- Result{Job: job, Err: err}
			continue
		}
		saveQueue <- job
	}
}

func saverWorker(id int, saveQueue <-chan ImageJob, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range saveQueue {
		log.Printf("[Saver %d] Saving %s", id, job.FileName)
		if err := saveImage(job.FileName); err != nil {
			results <- Result{Job: job, Err: err}
			continue
		}
		results <- Result{Job: job, Err: nil}
	}
}

func main() {

	imageFiles := []string{
		"wedding.jpg", "birthday.jpg", "vacation.jpg",
		"nature1.jpg", "nature2.jpg", "fashion.jpg",
	}

	numCPU := runtime.NumCPU()
	numProcessor := numCPU
	numSaver := numCPU * 2

	log.Printf("CPU cores: %d, Processor Workers: %d, Saver Workers: %d\n", numCPU, numProcessor, numSaver)

	jobQueue := make(chan ImageJob, numProcessor*2)
	saveQueue := make(chan ImageJob, numSaver*2)
	results := make(chan Result, len(imageFiles))

	var processorWg sync.WaitGroup
	var saverWg sync.WaitGroup

	for i := 0; i < numProcessor; i++ {
		processorWg.Add(1)
		go processorWorker(i, jobQueue, saveQueue, results, &processorWg)
	}

	for i := 0; i < numSaver; i++ {
		saverWg.Add(1)
		go saverWorker(i, saveQueue, results, &saverWg)
	}

	for _, img := range imageFiles {
		jobQueue <- ImageJob{FileName: img}
	}
	close(jobQueue)

	go func() {
		processorWg.Wait()
		close(saveQueue)
	}()

	go func() {
		saverWg.Wait()
		close(results)
	}()

	var failed int
	for res := range results {
		if res.Err != nil {
			log.Printf("❌ Failed: %s - %v", res.Job.FileName, res.Err)
			failed++
		} else {
			log.Printf("✅ Success: %s", res.Job.FileName)
		}
	}

	log.Printf("✅ All done. Total: %d, Failed: %d", len(imageFiles), failed)
}
