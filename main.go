package main

import (
	"database/sql"
	libroberto "github.com/TheTipo01/libRoberto"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/kkyr/fig"
	"math/rand"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type config struct {
	Token    string `fig:"token" validate:"required"`
	LogLevel string `fig:"loglevel" validate:"required"`
	Voice string `fig:"voice" validate:"required"`
}

var (
	// Discord bot token
	token string
	// Server
	server = make(map[string]*Server)
	// DB connection
	db *sql.DB
)

// DB parameters
const (
	dataSourceName = "./roberto.db"
	driverName     = "sqlite"
)

func init() {
	lit.LogLevel = lit.LogError

	var cfg config
	err := fig.Load(&cfg, fig.File("config.yml"))
	if err != nil {
		lit.Error(err.Error())
		return
	}

	libroberto.Voice = cfg.Voice
	token = cfg.Token

	// Set lit.LogLevel to the given value
	switch strings.ToLower(cfg.LogLevel) {
	case "logwarning", "warning":
		lit.LogLevel = lit.LogWarning

	case "loginformational", "informational":
		lit.LogLevel = lit.LogInformational

	case "logdebug", "debug":
		lit.LogLevel = lit.LogDebug
	}

	// Initialize rand
	rand.Seed(time.Now().Unix())

	// Database
	db, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		lit.Error("Error opening database connection, %s", err)
		return
	}

	execQuery(tblCustomCommands, db)

	loadCustomCommands(db)
}

func main() {
	if token == "" {
		lit.Error("No token provided. Please modify config.yml")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		lit.Error("Error creating Discord session: %s", err)
		return
	}

	// Add events handler
	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)

	// Add commands handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// We set the intents that we use
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		lit.Error("Error opening Discord session: %s", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	lit.Info("roberto is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()
}

func ready(s *discordgo.Session, _ *discordgo.Ready) {
	// Set the playing status.
	err := s.UpdateGameStatus(0, "Serving "+strconv.Itoa(len(s.State.Guilds))+" guilds!")
	if err != nil {
		lit.Error("Can't set status, %s", err)
	}

	// Checks for unused commands and deletes them
	if cmds, err := s.ApplicationCommands(s.State.User.ID, ""); err == nil {
		for _, c := range cmds {
			if commandHandlers[c.Name] == nil {
				_ = s.ApplicationCommandDelete(s.State.User.ID, "", c.ID)
				lit.Info("Deleted unused command %s", c.Name)
			}

			// Compare commands with the ones in commands, if they are different we re-create them
			for _, v := range commands {
				if c.Name == v.Name {
					if !isCommandEqual(c, v) {
						lit.Info("Command %s changed, re-creating", v.Name)
						_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
						if err != nil {
							lit.Error("Cannot create '%v' command: %v", v.Name, err)
						}
					}
					break
				}
			}
		}
	}

	// And add commands used
	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			lit.Error("Cannot create '%v' command: %v", v.Name, err)
		}
	}
}

func guildCreate(_ *discordgo.Session, e *discordgo.GuildCreate) {
	initializeServer(e.Guild.ID)
}
