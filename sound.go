package main

import (
	"encoding/binary"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io"
	"os"
	"time"
)

// playSound plays a file to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID, fileName string) {
	var opuslen int16

	file, err := os.Open("./temp/" + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}

	// Locks the mutex for the current server
	server[guildID].mutex.Lock()

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
		err = binary.Read(file, binary.LittleEndian, &opuslen)

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
		err = binary.Read(file, binary.LittleEndian, &InBuf)

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
			break
		}
	}

	// Close the file
	_ = file.Close()

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
func playSound2(fileName string, vc *discordgo.VoiceConnection, s *discordgo.Session) {
	var opuslen int16

	file, err := os.Open("./temp/" + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}

	// Start speaking.
	_ = vc.Speaking(true)

	// Channel to send ok messages
	c1 := make(chan string, 1)

	guildID := vc.GuildID
	channelID := vc.ChannelID

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

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
		err = binary.Read(file, binary.LittleEndian, &InBuf)

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

	// Close the file
	_ = file.Close()

	// Stop speaking
	_ = vc.Speaking(false)
}
