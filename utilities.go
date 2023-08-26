package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"math/rand"
	"strings"
	"time"
)

// findUserVoiceState finds the voice state of a user
func findUserVoiceState(s *discordgo.Session, guildID string, userID string) *discordgo.VoiceState {
	for _, g := range s.State.Guilds {
		if g.ID == guildID {
			for _, vs := range g.VoiceStates {
				if vs.UserID == userID {
					return vs
				}
			}
		}
	}
	return nil
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
func initializeServer(guildID string) {
	if server[guildID] == nil {
		server[guildID] = NewServer(guildID)
	}
}

// Sends embed as response to an interaction
func sendEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, c chan<- struct{}) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}}})
	if err != nil {
		lit.Error("InteractionRespond failed: %s", err)
		return
	}

	if c != nil {
		c <- struct{}{}
	}
}

// Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration) {
	sendEmbedInteraction(s, embed, i, nil)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

func sendEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string) *discordgo.Message {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		lit.Error("MessageSendEmbed failed: %s", err)
		return nil
	}

	return m
}

// joinVC joins the voice channel if not already joined, returns true if joined successfully
func joinVC(s *discordgo.Session, i *discordgo.Interaction, channelID string) bool {
	if server[i.GuildID].vc == nil {
		// Join the voice channel
		vc, err := s.ChannelVoiceJoin(i.GuildID, channelID, false, true)
		if err != nil {
			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, cantJoinVC).
				SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
			return false
		}
		server[i.GuildID].vc = vc
		server[i.GuildID].voiceChannel = channelID
	}
	return true
}

// Disconnects the bot from the voice channel
func quitVC(guildID string) {
	if server[guildID].queue.IsEmpty() && server[guildID].vc != nil {
		_ = server[guildID].vc.Disconnect()
		server[guildID].vc = nil
		server[guildID].voiceChannel = ""
	}
}

func deleteInteraction(s *discordgo.Session, i *discordgo.Interaction, c <-chan struct{}) {
	if c != nil {
		<-c
	}
	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}
