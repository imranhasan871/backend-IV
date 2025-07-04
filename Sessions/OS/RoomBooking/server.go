package main

import (
	"fmt"
	"sync"
	"time"
)

type Booking struct {
	UserID    int
	StartDate time.Time
	EndDate   time.Time
}

type Hotel struct {
	mu           sync.Mutex
	reservations map[int][]Booking // key: roomID, value: list of bookings
}

func NewHotel() *Hotel {
	return &Hotel{
		reservations: make(map[int][]Booking),
	}
}

// Check if requested dates overlap with existing bookings
func datesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

// BookRoom tries to book the room if no overlapping bookings exist.
// Uses locking to prevent race conditions.
func (h *Hotel) BookRoom(roomID int, userID int, startDate, endDate time.Time) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	bookings := h.reservations[roomID]
	for _, b := range bookings {
		if datesOverlap(startDate, endDate, b.StartDate, b.EndDate) {
			return fmt.Errorf("room %d already booked for overlapping dates", roomID)
		}
	}

	// No overlap, book the room
	h.reservations[roomID] = append(h.reservations[roomID], Booking{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})

	return nil
}

func main() {
	hotel := NewHotel()

	// Define a booking time window
	start := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 7, 5, 0, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup

	// Simulate multiple users trying to book the same room concurrently
	for user := 1; user <= 100; user++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			err := hotel.BookRoom(101, userID, start, end)
			if err != nil {
				fmt.Printf("User %d: failed to book room 101: %s\n", userID, err)
			} else {
				fmt.Printf("User %d: successfully booked room 101\n", userID)
			}
		}(user)
	}

	wg.Wait()
}
