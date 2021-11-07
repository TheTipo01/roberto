package main

import "sync"

// Server holds info about a guild
type Server struct {
	// Mutex for syncing requests
	mutex *sync.Mutex
	// Boolean for skipping
	stop bool
	// Custom commands
	customCommands map[string]string
}
