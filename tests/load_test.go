package tests

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoadSimulation(t *testing.T) {
	const (
		numWorkers        = 100
		requestsPerWorker = 10
	)

	var (
		successCount int64
		errorCount   int64
		wg           sync.WaitGroup
	)

	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				err := simulateAPIRequest()
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)
	total := numWorkers * requestsPerWorker

	fmt.Printf("Load Test Results:\n")
	fmt.Printf("Total Requests: %d\n", total)
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("RPS: %.2f\n", float64(total)/duration.Seconds())

	if errorCount > 0 {
		t.Fatalf("Load test failed with %d errors", errorCount)
	}
}

func simulateAPIRequest() error {
	time.Sleep(10 * time.Millisecond)
	return nil
}
