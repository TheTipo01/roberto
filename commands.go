package main

import (
	"github.com/TheTipo01/libRoberto"
	"github.com/TheTipo01/roberto/queue"
	"github.com/bwmarrin/discordgo"
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
			var n, j int64

			// If a number is given, we repeat the bestemmia n times
			if len(i.ApplicationCommandData().Options) > 0 {
				n = int64(i.ApplicationCommandData().Options[0].Value.(float64))
			} else {
				n = 1
			}

			bestemmie := make([]string, n)

			for ; j < n; j++ {
				bestemmie[j] = libroberto.Bestemmia()
			}

			playCommand(s, i, "Bestemmia", bestemmie...)
		},

		// Says text out lout
		"say": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			text := i.ApplicationCommandData().Options[0].StringValue()
			playCommand(s, i, "Say", libroberto.EmojiToDescription(text))
		},

		// Stops all commands
		"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].IsPlaying() {
					server[i.GuildID].Clear()
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Stop", "Stopped everything").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, "Nothing currently playing!").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Fakes train announcement
		"treno": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			trainAnnounce := libroberto.SearchAndGetTrain(i.ApplicationCommandData().Options[0].StringValue())
			if trainAnnounce == "" {
				trainAnnounce = "Nessun treno trovato, agagagaga!"
			}

			playCommand(s, i, "Treno", trainAnnounce)
		},

		// Says covid data out lout
		"covid": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			covid := libroberto.GetCovid()
			playCommand(s, i, "Covid", covid)
		},

		// Adds a custom command
		"addcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := addCommand(i.ApplicationCommandData().Options[0].StringValue(), i.ApplicationCommandData().Options[1].StringValue(), i.GuildID)
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successTitle, "Custom command added!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Removes a custom command
		"rmcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := removeCustom(i.ApplicationCommandData().Options[0].StringValue(), i.GuildID)
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successTitle, "Command removed successfully!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Select a random custom command
		"preghiera": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(server[i.GuildID].customCommands) > 0 {
				text := libroberto.EmojiToDescription(advancedReplace(advancedReplace(getRand(server[i.GuildID].customCommands), "<god>", libroberto.Gods), "<dict>", libroberto.Adjectives))
				playCommand(s, i, "Preghiera", text)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, "No custom commands available in this server! Add some with /addcustom").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Plays the custom command if it exists
		"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			command := i.ApplicationCommandData().Options[0].StringValue()
			if server[i.GuildID].customCommands[command] != "" {
				text := libroberto.EmojiToDescription(advancedReplace(advancedReplace(server[i.GuildID].customCommands[command], "<god>", libroberto.Gods), "<dict>", libroberto.Adjectives))
				playCommand(s, i, "Custom", text)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, "Command doesn't exist!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// List all of the custom commands for the server
		"listcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			message := ""

			for c := range server[i.GuildID].customCommands {
				message += c + ", "
			}

			if len(message) >= 2 {
				message = message[:len(message)-2]
			}

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Commands", message).
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*30)
		},

		// List all of the custom commands for the server
		"wikipedia": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			article := libroberto.EmojiToDescription(libroberto.GetWikipedia(i.ApplicationCommandData().Options[0].StringValue()))
			playCommand(s, i, "Wikipedia", article)
		},
	}
)

func playCommand(s *discordgo.Session, i *discordgo.InteractionCreate, title string, content ...string) {
	// Check if user is not in a voice channel
	if vs := findUserVoiceState(s, i.GuildID, i.Member.User.ID); vs != nil {
		c := make(chan struct{})
		go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Processing", ":)").SetColor(0x7289DA).MessageEmbed, i.Interaction, c)

		if joinVC(s, i.Interaction, vs.ChannelID) {
			elements := make([]queue.Element, len(content))

			for j, c := range content {
				cmds := libroberto.GenDCA(c)
				dcaOut, _ := cmds[2].StdoutPipe()

				elements[j] = queue.Element{
					Reader: dcaOut,
					BeforePlay: func() {
						libroberto.CmdsStart(cmds)
					},
					AfterPlay: func() {
						libroberto.CmdsKill(cmds)
						libroberto.CmdsWait(cmds)
					},
					Type:        title,
					Content:     c,
					TextChannel: i.ChannelID,
				}
			}

			server[i.GuildID].AddSong(false, elements...)
			go deleteInteraction(s, i.Interaction, c)
		}
	} else {
		sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
			SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
	}
}
