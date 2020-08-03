package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

// deleteMessage delete a message
func deleteMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println(m.Author.Username + ": " + m.Content)
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		fmt.Println("Can't delete message,", err)
	}
}

// findUserVoiceState finds the voicestate of a user
func findUserVoiceState(session *discordgo.Session, userid string) *discordgo.VoiceState {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs
			}
		}
	}
	return nil
}
