package main

import (
	"github.com/TheTipo01/libRoberto"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"os/exec"
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
					Name:        "custom-command",
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
					Name:        "custom-command",
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
					Name:        "custom-command",
					Description: "Command name to remove",
					Required:    true,
				},
			},
		},
		{
			Name:        "listcustom",
			Description: "Listes all custom command for the server",
		},
		{
			Name:        "wikipedia",
			Description: "Says wikipedia article out lout",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Wikipedia article",
					Required:    true,
				},
			},
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
			if len(i.ApplicationCommandData().Options) > 0 {
				var (
					cont int64
					n    = i.ApplicationCommandData().Options[0].IntValue()
					cmds []*exec.Cmd
				)

				for cont = 0; cont < n; cont++ {
					if server[vs.GuildID].stop {
						bstm := libroberto.Bestemmia()

						if cont == 0 {
							go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Bestemmia", bstm).
								SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)
						} else {
							go modfyInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Bestemmia", bstm).
								SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)
						}

						cmds = libroberto.GenDCA(strings.ToUpper(bstm))
						playSound2(vc, s, cmds)

						<-c
					} else {
						// Resets the stop boolean
						server[vs.GuildID].stop = true

						// Kill the processes, as we don't need to wait for them to finish
						libroberto.CmdsKill(cmds)
						break
					}
				}
			} else {
				// Else, we only do the command once
				bstm := libroberto.Bestemmia()

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Bestemmia", bstm).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound2(vc, s, libroberto.GenDCA(strings.ToUpper(bstm)))
				<-c
			}

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
				text := i.ApplicationCommandData().Options[0].StringValue()
				c := make(chan int)

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Say", text).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound(s, vs.GuildID, vs.ChannelID, libroberto.GenDCA(libroberto.EmojiToDescription(text)))

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

				trainAnnounce := libroberto.SearchAndGetTrain(i.ApplicationCommandData().Options[0].StringValue())
				if trainAnnounce == "" {
					trainAnnounce = "Nessun treno trovato, agagagaga!"
				}

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Treno", trainAnnounce).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound(s, vs.GuildID, vs.ChannelID, libroberto.GenDCA(trainAnnounce))

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
				covid := libroberto.GetCovid()

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Covid", covid).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, nil)

				playSound(s, vs.GuildID, vs.ChannelID, libroberto.GenDCA(covid))
			}
		},

		// Adds a custom command
		"addcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := addCommand(i.ApplicationCommandData().Options[0].StringValue(), i.ApplicationCommandData().Options[1].StringValue(), i.GuildID)
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
			err := removeCustom(i.ApplicationCommandData().Options[0].StringValue(), i.GuildID)
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
				text := libroberto.EmojiToDescription(advancedReplace(advancedReplace(getRand(server[i.GuildID].customCommands), "<god>", libroberto.Gods), "<dict>", libroberto.Adjectives))
				c := make(chan int)

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Preghiera", text).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound(s, vs.GuildID, vs.ChannelID, libroberto.GenDCA(text))

				<-c
				err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
				if err != nil {
					lit.Error("InteractionResponseDelete failed: %s", err.Error())
				}
			}
		},

		// Plays the custom command if it exist
		"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			command := i.ApplicationCommandData().Options[0].StringValue()
			if server[i.GuildID].customCommands[command] != "" {
				vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
				if vs != nil {
					text := libroberto.EmojiToDescription(advancedReplace(advancedReplace(server[i.GuildID].customCommands[command], "<god>", libroberto.Gods), "<dict>", libroberto.Adjectives))
					c := make(chan int)

					go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Custom", text).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

					playSound(s, vs.GuildID, vs.ChannelID, libroberto.GenDCA(text))

					<-c
					err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
					if err != nil {
						lit.Error("InteractionResponseDelete failed: %s", err.Error())
					}
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

		// List all of the custom commands for the server
		"wikipedia": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Member.User.ID, i.GuildID)
			if vs != nil {
				c := make(chan int)

				article := libroberto.EmojiToDescription(libroberto.GetWikipedia(i.ApplicationCommandData().Options[0].StringValue()))

				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Wikipedia", article).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

				playSound(s, vs.GuildID, vs.ChannelID, libroberto.GenDCA(article))

				<-c
				err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
				if err != nil {
					lit.Error("InteractionResponseDelete failed: %s", err.Error())
				}
			}
		},
	}
)
