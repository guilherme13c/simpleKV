package simplekv_test

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"simpleKV/client"
	"simpleKV/server"
	"simpleKV/server/store"
	"testing"
	"time"
)

func TestSystem(t *testing.T) {
	// Start the server
	go func() {
		store := store.NewStore(4, 4096)
		server := server.NewServer(":6379", store)

		err := server.Run()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	time.Sleep(time.Second * 3) // Give the server a moment to start

	// Connect to the server
	c, err := client.NewClient("localhost:6379")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer c.Close()

	// Test SET command
	if err := c.Set("key1", "hello"); err != nil {
		t.Errorf("SET command failed: %v", err)
	}

	// Test GET command
	val, err := c.Get("key1")
	if err != nil || val != "hello" {
		t.Errorf("GET command failed: got %v, expected 'hello'", val)
	}

	// Test DEL command
	if err := c.Del("key1"); err != nil {
		t.Errorf("DEL command failed: %v", err)
	}
	val, err = c.Get("key1")
	if err != nil || val != nil {
		t.Errorf("GET after DEL failed: expected nil, got %v", val)
	}

	// Test INFO command
	if err := c.Info(); err != nil {
		t.Errorf("INFO command failed: %v", err)
	}

	// Test COMMAND command
	if err := c.Command("DOCS"); err != nil {
		t.Errorf("COMMAND command failed: %v", err)
	}

	// Test SCAN command
	if err := c.Set("user:1", "John"); err != nil {
		t.Errorf("SET failed: %v", err)
	}
	if err := c.Set("user:2", "Jane"); err != nil {
		t.Errorf("SET failed: %v", err)
	}
	if err := c.Set("item:1", "Book"); err != nil {
		t.Errorf("SET failed: %v", err)
	}

	keys, nextCursor, err := c.Scan(0, regexp.MustCompile("user:*"), 10)
	if err != nil {
		t.Errorf("SCAN failed: %v", err)
	}
	if nextCursor != 0 || len(keys) != 2 {
		t.Errorf("SCAN returned unexpected result: cursor=%d, keys=%v", nextCursor, keys)
	}

	// Test data persistence
	if err := c.Set("persistentKey", "persistentValue"); err != nil {
		t.Errorf("Failed to set persistent key: %v", err)
	}
	c.Close()

	// Restart the server to test persistence
	if err := os.Remove("dump.rdb"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove dump.rdb: %v", err)
	}

	go func() {
		store := store.NewStore(4, 4096)
		server := server.NewServer(":6379", store)

		err := server.Run()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	time.Sleep(time.Second * 3)

	// Reconnect
	c, err = client.NewClient("localhost:6379")
	if err != nil {
		t.Fatalf("Failed to reconnect to server: %v", err)
	}
	defer c.Close()

	// Check if data is persisted
	val, err = c.Get("persistentKey")
	if err != nil || val != "persistentValue" {
		t.Errorf("Persistence failed: got %v, expected 'persistentValue'", val)
	}

	fmt.Println("System test completed successfully.")
}
