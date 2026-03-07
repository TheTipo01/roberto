package main

import (
	"sync/atomic"

	"github.com/TheTipo01/roberto/queue"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

// Server holds info about a guild
type Server struct {
	// Channel for skipping
	skip chan struct{}
	// Custom commands
	customCommands map[string]string
	// Voice connection
	vc voice.Conn
	// Voice channel
	voiceChannel *snowflake.ID
	// Queue
	queue queue.Queue
	// Whether the job scheduler has started
	started atomic.Bool
	// Whether to clear the queue
	clear atomic.Bool
	// Guild ID
	guildID snowflake.ID
}
