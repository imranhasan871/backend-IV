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
	lock         sync.Mutex          // protects access to maps
	reservations map[int][]Booking   // roomID → list of bookings
	roomLocks    map[int]*sync.Mutex // roomID → mutex for that room
}

func NewHotel() *Hotel {
	return &Hotel{
		reservations: make(map[int][]Booking),
		roomLocks:    make(map[int]*sync.Mutex),
	}
}

func (h *Hotel) getRoomLock(roomID int) *sync.Mutex {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, exists := h.roomLocks[roomID]; !exists {
		h.roomLocks[roomID] = &sync.Mutex{}
	}
	return h.roomLocks[roomID]
}

func (h *Hotel) BookRoom(roomID int, userID int, startDate, endDate time.Time) error {
	roomMu := h.getRoomLock(roomID)
	roomMu.Lock()
	defer roomMu.Unlock()

	bookings := h.reservations[roomID]
	for _, b := range bookings {
		if datesOverlap(startDate, endDate, b.StartDate, b.EndDate) {
			return fmt.Errorf("room %d already booked for overlapping dates", roomID)
		}
	}

	h.reservations[roomID] = append(h.reservations[roomID], Booking{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})

	return nil
}

func datesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

func main() {
	hotel := NewHotel()
	var wg sync.WaitGroup

	start := time.Now()
	end := start.Add(2 * time.Hour)

	// Simulate 100 users trying to book the same room
	for user := 1; user <= 100; user++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			err := hotel.BookRoom(101, userID, start, end)
			if err != nil {
				fmt.Printf("User %d: ❌ failed to book room 101: %s\n", userID, err)
			} else {
				fmt.Printf("User %d: ✅ successfully booked room 101\n", userID)
			}
		}(user)
	}

	// Simulate 100 users trying to book a different room
	for user := 101; user <= 200; user++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			err := hotel.BookRoom(102, userID, start, end)
			if err != nil {
				fmt.Printf("User %d: ❌ failed to book room 102: %s\n", userID, err)
			} else {
				fmt.Printf("User %d: ✅ successfully booked room 102\n", userID)
			}
		}(user)
	}

	wg.Wait()
	fmt.Println("🏁 All booking attempts finished.")
}
