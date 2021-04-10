package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"strings"
	"time"
)

var (
	// Commands
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "say",
			Description: "Says text out loud",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "Text to say out loud",
					Required:    true,
				},
			},
		},
		{
			Name:        "bestemmia",
			Description: "Generates a bestemmia n times",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "n",
					Description: "Number of times to generate bestemmia",
					Required:    false,
				},
			},
		},
		{
			Name:        "treno",
			Description: "Fakes train announcement given it's number",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "train",
					Description: "Train number",
					Required:    true,
				},
			},
		},
		{
			Name:        "covid",
			Description: "Says covid data out loud for current day in Italy",
		},
		{
			Name:        "preghiera",
			Description: "Randomly select a custom command",
		},
		{
			Name:        "stop",
			Description: "Stops every command",
		},
		{
			Name:        "addcustom",
			Description: "Creates a custom command. <god> will be replace with a random god and <dict> with an adjective",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "customCommand",
					Description: "Command name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "Text to say out loud",
					Required:    true,
				},
			},
		},
		{
			Name:        "rmcustom",
			Description: "Removes a custom command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "customCommand",
					Description: "Command name to remove",
					Required:    true,
				},
			},
		},
		{
			Name:        "custom",
			Description: "Calls a custom command. Use /listcustom for a list",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "customCommand",
					Description: "Command name to remove",
					Required:    true,
				},
			},
		},
		{
			Name:        "listcustom",
			Description: "Listes all custom command for the server",
		},
	}

	// Handler
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// Generates random bestemmie
		"bestemmia": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			c := make(chan int)

			// Locks the mutex for the current server
			server[vs.GuildID].mutex.Lock()

			// Join the provided voice channel.
			vc, err := s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
			if err != nil {
				lit.Error("Can't connect to voice channel, %s", err)
				server[vs.GuildID].mutex.Unlock()
				return
			}

			// If a number is given, we repeat the bestemmia n times
			if len(i.Data.Options) > 0 {
				var (
					cont uint64
					n    = i.Data.Options[0].UintValue()
				)

				for cont = 0; cont < n; cont++ {
					if server[vs.GuildID].stop {
						bstm := bestemmia()

						if cont == 0 {
							go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Bestemmia", bstm).
								SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)
						} else {
							go modfyInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Bestemmia", bstm).
								SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)
						}

						playSound2(genAudio(strings.ToUpper(bstm)), vc, s)

						<-c
					} else {
						// Resets the stop boolean
						server[vs.GuildID].stop = true
						break
					}
				}
			} else {
				// Else, we only do the command once
				bstm := bestemmia()

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Bestemmia", bstm).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound2(genAudio(strings.ToUpper(bstm)), vc, s)
			}

			<-c
			// Deletes interaction as we have finished
			err = s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			if err != nil {
				lit.Error("InteractionResponseDelete failed: %s", err.Error())
			}

			// Disconnect from the provided voice channel.
			err = vc.Disconnect()
			if err != nil {
				lit.Error("Can't disconnect from voice channel, %s", err)
			}

			// Releases the mutex lock for the server
			server[vs.GuildID].mutex.Unlock()
		},

		// Says text out lout
		"say": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
			if vs != nil {
				text := i.Data.Options[0].StringValue()
				c := make(chan int)

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Say", text).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound(s, vs.GuildID, vs.ChannelID, genAudio(emojiToDescription(text)))

				<-c
				err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
				if err != nil {
					lit.Error("InteractionResponseDelete failed: %s", err.Error())
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Stops all commands
		"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			server[i.GuildID].stop = false

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Stop", "Stopped").
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
		},

		// Fakes train announcement
		"treno": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
			if vs != nil {
				c := make(chan int)

				trainAnnounce := searchAndGetTrain(i.Data.Options[0].StringValue())
				if trainAnnounce == "" {
					trainAnnounce = "Nessun treno trovato, agagagaga!"
				}

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Treno", trainAnnounce).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound(s, vs.GuildID, vs.ChannelID, genAudio(trainAnnounce))

				<-c
				err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
				if err != nil {
					lit.Error("InteractionResponseDelete failed: %s", err.Error())
				}
			}
		},

		// Says covid data out lout
		"covid": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
			if vs != nil {
				covid := getCovid()

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Covid", covid).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, nil)

				playSound(s, vs.GuildID, vs.ChannelID, genAudio(covid))
			}
		},

		// Adds a custom command
		"addcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := addCommand(i.Data.Options[0].StringValue(), i.Data.Options[1].StringValue(), i.GuildID)
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Successful", "Custom command added!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Removes a custom command
		"rmcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := removeCustom(i.Data.Options[0].StringValue(), i.GuildID)
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Successful", "Command removed successfully!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Select a random custom command
		"preghiera": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
			if vs != nil {
				text := advancedReplace(advancedReplace(getRand(server[i.GuildID].customCommands), "<god>", gods), "<dict>", adjectives)
				c := make(chan int)

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Preghiera", text).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound(s, vs.GuildID, vs.ChannelID, genAudio(text))

				<-c
				err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
				if err != nil {
					lit.Error("InteractionResponseDelete failed: %s", err.Error())
				}
			}
		},

		// Plays the custom command if it exist
		"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			command := i.Data.Options[0].StringValue()
			if server[i.GuildID].customCommands[command] != "" {
				vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
				if vs != nil {
					playSound(s, vs.GuildID, vs.ChannelID, genAudio(advancedReplace(advancedReplace(server[i.GuildID].customCommands[command], "<god>", gods), "<dict>", adjectives)))
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Command doesn't exist!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// List all of the custom commands for the server
		"listcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			message := ""

			for c := range server[i.GuildID].customCommands {
				message += c + ", "
			}

			message = message[:len(message)-2]

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Commands", message).
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*30)
		},
	}
)
