package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Document represents the collaboratively edited document
type Document struct {
	mu      sync.RWMutex // now using RWMutex
	Content string
	Version int
}

// SaveRequest holds user-provided data for a save
type SaveRequest struct {
	UserID  string
	Content string
	Version int
}

// GetSnapshot returns a copy of the document content and version
func (d *Document) GetSnapshot() (string, int) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Content, d.Version
}

// TrySave attempts to save if the version matches
func (d *Document) TrySave(req SaveRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if req.Version != d.Version {
		return fmt.Errorf("version conflict: current is %d, you provided %d", d.Version, req.Version)
	}

	d.Content = req.Content
	d.Version++
	return nil
}

// Merge appends user content with a marker when version conflict happens
func (d *Document) Merge(userContent string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Content += "\n<<MERGED>>\n" + userContent
	d.Version++
}

func main() {
	doc := &Document{
		Content: "Initial content.",
		Version: 1,
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// Show current state
		content, version := doc.GetSnapshot()
		fmt.Println("\n=== Collaborative Document ===")
		fmt.Printf("Current Version: %d\n", version)
		fmt.Printf("Current Content:\n%s\n", content)

		// Get user info
		fmt.Print("\nEnter your name (e.g., Alice or Bob): ")
		userID, _ := reader.ReadString('\n')
		userID = strings.TrimSpace(userID)

		fmt.Printf("[User %s] What version do you think the document is? ", userID)
		var userVersion int
		fmt.Scanf("%d\n", &userVersion)

		fmt.Printf("[User %s] Enter your new content:\n> ", userID)
		userContent, _ := reader.ReadString('\n')
		userContent = strings.TrimSpace(userContent)

		// Try to save
		req := SaveRequest{
			UserID:  userID,
			Content: userContent,
			Version: userVersion,
		}

		err := doc.TrySave(req)
		if err != nil {
			fmt.Printf("⚠️  Save failed for %s: %v\n", userID, err)
			fmt.Print("Would you like to (M)erge or (R)etry? ")
			choice, _ := reader.ReadString('\n')
			choice = strings.ToUpper(strings.TrimSpace(choice))

			switch choice {
			case "M":
				doc.Merge(userContent)
				fmt.Println("✅ Merge applied successfully.")
			case "R":
				fmt.Println("🔁 Let's retry...")
				continue
			default:
				fmt.Println("❌ Invalid choice. Skipping this edit...")
			}
		} else {
			fmt.Printf("✅ Save successful. Version updated.\n")
		}
	}
}
