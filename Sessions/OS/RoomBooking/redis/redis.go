package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

var (
	ctx  = context.Background()
	room = "room-101"
)

func bookRoom(locker *redislock.Client, user string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Try to acquire lock on the room
	lock, err := locker.Obtain(ctx, room, 5*time.Second, nil)
	if err == redislock.ErrNotObtained {
		fmt.Printf("%s: could not book %s (already locked)\n", user, room)
		return
	} else if err != nil {
		log.Fatalf("lock error: %v", err)
	}
	defer lock.Release(ctx)

	// Simulate booking process
	fmt.Printf("%s: successfully booked %s!\n", user, room)
	time.Sleep(2 * time.Second) // Simulate processing time
}

func main() {
	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // adjust if needed
	})
	defer client.Close()

	// Create a Redlock client
	locker := redislock.New(client)

	var wg sync.WaitGroup
	users := []string{"Alice", "Bob", "Charlie"}

	// Simulate concurrent booking attempts
	for _, user := range users {
		wg.Add(1)
		go bookRoom(locker, user, &wg)
	}

	wg.Wait()
}
