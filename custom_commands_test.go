package main

import (
	"sync"
	"testing"

	"github.com/disgoorg/snowflake/v2"
)

// Ensures concurrent reads and writes to the custom commands map are safe.
// Run with -race to detect data races.
func TestCustomCommandsConcurrency(t *testing.T) {
	srv := NewServer(snowflake.ID(1))

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(5)

		go func() {
			defer wg.Done()
			srv.SetCustomCommand("command", "text")
		}()

		go func() {
			defer wg.Done()
			srv.GetCustomCommand("command")
		}()

		go func() {
			defer wg.Done()
			srv.DeleteCustomCommand("command")
		}()

		go func() {
			defer wg.Done()
			srv.ListCustomCommands()
		}()

		go func() {
			defer wg.Done()
			srv.CustomCommandsCount()
			srv.GetCustomCommands()
		}()
	}

	wg.Wait()
}
