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

// Initialize server for a given guildID if its nil
func initializeServer(guildID snowflake.ID) {
	if server[guildID] == nil {
		server[guildID] = NewServer(guildID)
	}
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
	if server[guildID].vc == nil {
		// Create the voice connection
		server[guildID].vc = e.Client().VoiceManager.CreateConn(guildID)
	}

	if server[guildID].voiceChannel == nil {
		// Join the voice channel
		err := server[guildID].vc.Open(context.TODO(), channelID, false, true)
		if err != nil {
			sendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(errorTitle, cantJoinVC, false).
				SetColor(0x7289DA).Build(), e, time.Second*5)
			return false
		}

		server[guildID].voiceChannel = &channelID
	}

	return true
}

// Disconnects the bot from the voice channel
func quitVC(guildID snowflake.ID) {
	if server[guildID].queue.IsEmpty() && server[guildID].voiceChannel != nil {
		server[guildID].vc.Close(context.TODO())
		server[guildID].voiceChannel = nil
		server[guildID].vc = nil
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
