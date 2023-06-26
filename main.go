package main

import (
	"database/sql"
	libroberto "github.com/TheTipo01/libRoberto"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/kkyr/fig"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type config struct {
	Token    string `fig:"token" validate:"required"`
	LogLevel string `fig:"loglevel" validate:"required"`
	Voice    string `fig:"voice" validate:"required"`
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
	dg.AddHandler(guildDelete)

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

	// Register commands
	lit.Info("Registering commands, this will take a while...")
	for _, v := range commands {
		_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
		if err != nil {
			lit.Error("Can't register command %s: %s", v.Name, err.Error())
		}
	}

	// Wait here until CTRL-C or another term signal is received.
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
}

func guildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	initializeServer(e.Guild.ID)
	ready(s, nil)
}

func guildDelete(s *discordgo.Session, _ *discordgo.GuildDelete) {
	ready(s, nil)
}
