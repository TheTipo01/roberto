package main

import (
	"encoding/binary"
	libroberto "github.com/TheTipo01/libRoberto"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io"
	"os/exec"
	"time"
)

// playSound plays a file to the provided channel.
func playSound(s *discordgo.Session, guildID string, channelID string, cmds []*exec.Cmd) {
	var opuslen int16

	// Locks the mutex for the current server
	server[guildID].mutex.Lock()

	pipe, _ := cmds[2].StdoutPipe()
	libroberto.CmdsStart(cmds)

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		server[guildID].mutex.Unlock()
		return
	}

	// Start speaking.
	_ = vc.Speaking(true)
	server[guildID].stop = true

	// Channel to send ok messages
	c1 := make(chan string, 1)

	for {
		// Read opus frame length from dca file.
		err = binary.Read(pipe, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(pipe, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Stream data to discord
		if server[guildID].stop {
			// Send data in a goroutine
			go func() {
				vc.OpusSend <- InBuf
				c1 <- "ok"
			}()

			// So if the bot gets disconnect/moved we can rejoin the original channel and continue playing songs
			select {
			case <-c1:
				break
			case <-time.After(time.Second / 3):
				vc, _ = s.ChannelVoiceJoin(guildID, channelID, false, true)
			}
		} else {
			// Kill the processes, as we don't need to wait for them to finish
			libroberto.CmdsKill(cmds)
			break
		}
	}

	// If the sound is skipped, we kill the processes, so this isn't needed
	if !server[guildID].stop {
		libroberto.CmdsWait(cmds)
	}

	// Resets the stop boolean
	server[guildID].stop = true

	// Stop speaking
	_ = vc.Speaking(false)

	// Disconnect from the provided voice channel.
	err = vc.Disconnect()
	if err != nil {
		lit.Error("Can't disconnect from voice channel, %s", err)
		return
	}

	// Releases the mutex lock for the server
	server[guildID].mutex.Unlock()
}

// playSound2 plays a file to the provided channel given a voice connection.
func playSound2(vc *discordgo.VoiceConnection, s *discordgo.Session, cmds []*exec.Cmd) {
	var opuslen int16
	var err error

	pipe, _ := cmds[2].StdoutPipe()
	libroberto.CmdsStart(cmds)

	// Start speaking.
	_ = vc.Speaking(true)

	// Channel to send ok messages
	c1 := make(chan string, 1)

	guildID := vc.GuildID
	channelID := vc.ChannelID

	for {
		// Read opus frame length from dca file.
		err = binary.Read(pipe, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(pipe, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Stream data to discord
		// Send data in a goroutine
		go func() {
			vc.OpusSend <- InBuf
			c1 <- "ok"
		}()

		// So if the bot gets disconnect/moved we can rejoin the original channel and continue playing songs
		select {
		case <-c1:

		case <-time.After(time.Second / 3):
			vc, _ = s.ChannelVoiceJoin(guildID, channelID, false, true)
		}

	}

	libroberto.CmdsWait(cmds)

	// Stop speaking
	_ = vc.Speaking(false)
}
