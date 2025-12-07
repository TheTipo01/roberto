package main

import (
	"sync/atomic"

	"github.com/TheTipo01/roberto/queue"
	"github.com/bwmarrin/discordgo"
)

// Server holds info about a guild
type Server struct {
	// Channel for skipping
	skip chan struct{}
	// Custom commands
	customCommands map[string]string
	// Voice connection
	vc *discordgo.VoiceConnection
	// Voice channel
	voiceChannel string
	// Queue
	queue queue.Queue
	// Whether the job scheduler has started
	started atomic.Bool
	// Whether to clear the queue
	clear atomic.Bool
	// Guild ID
	guildID string
}
