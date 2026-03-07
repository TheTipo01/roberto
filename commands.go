package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/TheTipo01/libRoberto"
	"github.com/TheTipo01/roberto/queue"
	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var (
	// Commands
	commands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "say",
			Description: "Says text out loud",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "text",
					Description: "Text to say out loud",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "bestemmia",
			Description: "Generates a bestemmia n times",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "n",
					Description: "Number of times to generate bestemmia",
					Required:    false,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "treno",
			Description: "Fakes train announcement given it's number",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "train",
					Description: "Train number",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "covid",
			Description: "Says covid data out loud for current day in Italy",
		},
		discord.SlashCommandCreate{
			Name:        "preghiera",
			Description: "Randomly select a custom command",
		},
		discord.SlashCommandCreate{
			Name:        "stop",
			Description: "Stops every command",
		},
		discord.SlashCommandCreate{
			Name:        "addcustom",
			Description: "Creates a custom command. <god> will be replace with a random god and <dict> with an adjective",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "custom-command",
					Description: "Command name",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "text",
					Description: "Text to say out loud",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "rmcustom",
			Description: "Removes a custom command",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "custom-command",
					Description: "Command name to remove",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "custom",
			Description: "Calls a custom command. Use /listcustom for a list",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "custom-command",
					Description: "Command name to remove",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "listcustom",
			Description: "Listes all custom command for the server",
		},
		discord.SlashCommandCreate{
			Name:        "wikipedia",
			Description: "Says wikipedia article out lout",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "link",
					Description: "Wikipedia article",
					Required:    true,
				},
			},
		},
	}

	// Handler
	commandHandlers = map[string]func(e *events.ApplicationCommandInteractionCreate){
		// Generates random bestemmie
		"bestemmia": func(e *events.ApplicationCommandInteractionCreate) {
			var n, j int

			// If a number is given, we repeat the bestemmia n times
			if option, ok := e.SlashCommandInteractionData().OptInt("n"); ok {
				n = option
			} else {
				n = 1
			}

			bestemmie := make([]string, n)

			for ; j < n; j++ {
				bestemmie[j] = libroberto.Bestemmia()
			}

			playCommand(e, "Bestemmia", bestemmie...)
		},

		// Says text out lout
		"say": func(e *events.ApplicationCommandInteractionCreate) {
			text := e.SlashCommandInteractionData().String("text")
			playCommand(e, "Say", libroberto.EmojiToDescription(text))
		},

		// Stops all commands
		"stop": func(e *events.ApplicationCommandInteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(e.Client(), e.Member().GuildID, e.Member().User.ID) != nil {
				if server[*e.GuildID()].IsPlaying() {
					server[*e.GuildID()].Clear()
					sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField("Stop", "Stopped everything", false).
						WithColor(0x7289DA), e, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, "Nothing currently playing!", false).
						WithColor(0x7289DA), e, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, notInVC, false).
					WithColor(0x7289DA), e, time.Second*5)
			}
		},

		// Fakes train announcement
		"treno": func(e *events.ApplicationCommandInteractionCreate) {
			trainAnnounce := libroberto.SearchAndGetTrain(e.SlashCommandInteractionData().String("train"))
			if trainAnnounce == "" {
				trainAnnounce = "Nessun treno trovato, agagagaga!"
			}

			playCommand(e, "Treno", trainAnnounce)
		},

		// Says covid data out lout
		"covid": func(e *events.ApplicationCommandInteractionCreate) {
			covid := libroberto.GetCovid()
			playCommand(e, "Covid", covid)
		},

		// Adds a custom command
		"addcustom": func(e *events.ApplicationCommandInteractionCreate) {
			err := addCommand(e.SlashCommandInteractionData().String("custom-command"), e.SlashCommandInteractionData().String("text"), *e.GuildID())
			if err != nil {
				sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, err.Error(), false).
					WithColor(0x7289DA), e, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(successTitle, "Custom command added!", false).
					WithColor(0x7289DA), e, time.Second*5)
			}
		},

		// Removes a custom command
		"rmcustom": func(e *events.ApplicationCommandInteractionCreate) {
			err := removeCustom(e.SlashCommandInteractionData().String("custom-command"), *e.GuildID())
			if err != nil {
				sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, err.Error(), false).
					WithColor(0x7289DA), e, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(successTitle, "Command removed successfully!", false).
					WithColor(0x7289DA), e, time.Second*5)
			}
		},

		// Select a random custom command
		"preghiera": func(e *events.ApplicationCommandInteractionCreate) {
			if len(server[*e.GuildID()].customCommands) > 0 {
				text := libroberto.EmojiToDescription(advancedReplace(advancedReplace(getRand(server[*e.GuildID()].customCommands), "<god>", libroberto.Gods), "<dict>", libroberto.Adjectives))
				playCommand(e, "Preghiera", text)
			} else {
				sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, "No custom commands available in this server! Add some with /addcustom", false).
					WithColor(0x7289DA), e, time.Second*5)
			}
		},

		// Plays the custom command if it exists
		"custom": func(e *events.ApplicationCommandInteractionCreate) {
			command := e.SlashCommandInteractionData().String("custom-command")
			if server[*e.GuildID()].customCommands[command] != "" {
				text := libroberto.EmojiToDescription(advancedReplace(advancedReplace(server[*e.GuildID()].customCommands[command], "<god>", libroberto.Gods), "<dict>", libroberto.Adjectives))
				playCommand(e, "Custom", text)
			} else {
				sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, "Command doesn't exist!", false).
					WithColor(0x7289DA), e, time.Second*5)
			}
		},

		// List all of the custom commands for the server
		"listcustom": func(e *events.ApplicationCommandInteractionCreate) {
			message := ""

			for c := range server[*e.GuildID()].customCommands {
				message += c + ", "
			}

			if len(message) >= 2 {
				message = message[:len(message)-2]
			}

			sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField("Commands", message, false).
				WithColor(0x7289DA), e, time.Second*30)
		},

		// List all of the custom commands for the server
		"wikipedia": func(e *events.ApplicationCommandInteractionCreate) {
			link := e.SlashCommandInteractionData().String("link")
			article := libroberto.EmojiToDescription(libroberto.GetWikipedia(link))
			playCommand(e, "Wikipedia", article)
		},
	}
)

func playCommand(e *events.ApplicationCommandInteractionCreate, title string, content ...string) {
	// Check if user is not in a voice channel
	if vs := findUserVoiceState(e.Client(), e.Member().GuildID, e.Member().User.ID); vs != nil {
		c := make(chan struct{})
		go sendEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField("Processing", ":)", false).WithColor(0x7289DA), e, c)

		if joinVC(e, *vs.ChannelID, vs.GuildID) {
			elements := make([]queue.Element, len(content))

			for j, c := range content {
				var dcaOut io.ReadCloser
				cmds := libroberto.GenDCA(c)

				if restRoberto != "" {
					// Delete all the commands except the last one
					cmds = cmds[1:3]
					dcaOut, _ = cmds[1].StdoutPipe()

					// Setup query parameters
					endpoint, err := url.Parse(restRoberto)
					if err != nil {
						log.Fatal(err)
					}

					queryParams := url.Values{}
					queryParams.Set("token", restRobertoToken)
					queryParams.Set("text", c)
					queryParams.Set("voice", libroberto.Voice)

					endpoint.RawQuery = queryParams.Encode()
					resp, err := http.Get(endpoint.String())
					if err != nil {
						lit.Error("Error calling restRoberto: %s", err.Error())
						continue
					}

					// Get the response and give it to the first command
					cmds[0].Stdin = resp.Body
				} else {
					dcaOut, _ = cmds[2].StdoutPipe()
				}

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
					TextChannel: e.Channel().ID(),
				}
			}

			server[vs.GuildID].AddSong(false, elements...)
			go deleteInteraction(e, c)
		}
	} else {
		sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(errorTitle, notInVC, false).
			WithColor(0x7289DA), e, time.Second*5)
	}
}
