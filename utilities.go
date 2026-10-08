package main

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

// findUserVoiceState finds user current voice channel
func findUserVoiceState(s *bot.Client, guildID, userID snowflake.ID) *discord.VoiceState {
	v, found := s.Caches.VoiceState(guildID, userID)

	if !found {
		return nil
	}

	return &v
}

// advancedReplace returns src string with every instance of toReplace with a random item from a
func advancedReplace(src string, toReplace string, a []string) string {
	var dst = src

	for i := 0; i < strings.Count(src, toReplace); i++ {
		dst = strings.Replace(dst, toReplace, a[rand.Intn(len(a))], 1)
	}

	return dst
}

// Returns a random value from a map of string
func getRand(a map[string]string) string {
	// produce a pseudo-random number between 0 and len(a)-1
	i := int(float32(len(a)) * rand.Float32())
	for _, v := range a {
		if i == 0 {
			return v
		}
		i--
	}
	panic("impossible")
}

// getServer returns the server manager for the given guild, creating it if it
// doesn't exist yet. It is safe to call concurrently from multiple goroutines.
func getServer(guildID snowflake.ID) *Server {
	serverMutex.RLock()
	srv := server[guildID]
	serverMutex.RUnlock()

	if srv != nil {
		return srv
	}

	serverMutex.Lock()
	defer serverMutex.Unlock()

	if srv = server[guildID]; srv == nil {
		srv = NewServer(guildID)
		server[guildID] = srv
	}

	return srv
}

// Initialize server for a given guildID if its nil
func initializeServer(guildID snowflake.ID) {
	getServer(guildID)
}

// guildCount returns the number of guilds currently tracked
func guildCount() int {
	serverMutex.RLock()
	defer serverMutex.RUnlock()

	return len(server)
}

// Sends embed as response to an interaction
func sendEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, c chan<- struct{}) {
	err := e.CreateMessage(discord.NewMessageCreate().AddEmbeds(embed))
	if err != nil {
		lit.Error("InteractionRespond failed: %s", err)
		return
	}

	if c != nil {
		c <- struct{}{}
	}
}

// Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, wait time.Duration) {
	sendEmbedInteraction(embed, e, nil)

	time.Sleep(wait)

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

func sendEmbed(c *bot.Client, embed discord.Embed, txtChannel snowflake.ID) *discord.Message {
	m, err := c.Rest.CreateMessage(txtChannel, discord.NewMessageCreate().AddEmbeds(embed))
	if err != nil {
		lit.Error("sendEmbed failed: %s", err)
		return nil
	}

	return m
}

// joinVC joins the voice channel if not already joined, returns true if joined successfully
func joinVC(e *events.ApplicationCommandInteractionCreate, channelID, guildID snowflake.ID) bool {
	srv := getServer(guildID)

	if srv.vc == nil {
		// Create the voice connection
		srv.vc = e.Client().VoiceManager.CreateConn(guildID)
	}

	if srv.voiceChannel == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			// Join the voice channel
			errCh <- srv.vc.Open(ctx, channelID, false, true)
		}()

		var err error
		select {
		case err = <-errCh:
		case <-ctx.Done():
			err = ctx.Err()
		}

		if err != nil {
			sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, cantJoinVC, false).
				WithColor(0x7289DA), e, time.Second*5)
			return false
		}

		srv.voiceChannel = &channelID
	}

	return true
}

// Disconnects the bot from the voice channel
func quitVC(guildID snowflake.ID) {
	srv := getServer(guildID)

	if srv.queue.IsEmpty() && srv.voiceChannel != nil {
		srv.vc.Close(context.TODO())
		srv.voiceChannel = nil
		srv.vc = nil
	}
}

func deleteInteraction(e *events.ApplicationCommandInteractionCreate, c <-chan struct{}) {
	if c != nil {
		<-c
	}

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		lit.Error("DeleteInteractionResponse failed: %s", err)
		return
	}
}
