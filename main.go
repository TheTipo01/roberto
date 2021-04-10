package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/spf13/viper"
	"io/ioutil"
	"math/rand"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	// Discord bot token
	token string
	// Server
	server = make(map[string]*Server)
	// Array of adjectives
	adjectives []string
	// Gods
	gods = []string{"Dio", "Gesù", "Madonna"}
	// Emoji replacer
	emoji = *emojiReplacer()
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

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			lit.Error("Config file not found! See example_config.yml")
			return
		}
	} else {
		// Config file found
		token = viper.GetString("token")

		// Set lit.LogLevel to the given value
		switch strings.ToLower(viper.GetString("loglevel")) {
		case "logerror", "error":
			lit.LogLevel = lit.LogError
			break
		case "logwarning", "warning":
			lit.LogLevel = lit.LogWarning
			break
		case "loginformational", "informational":
			lit.LogLevel = lit.LogInformational
			break
		case "logdebug", "debug":
			lit.LogLevel = lit.LogDebug
			break
		}

		initializeAdjectives()

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
		if h, ok := commandHandlers[i.Data.Name]; ok {
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
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
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

// Reads adjectives
func initializeAdjectives() {
	foo, _ := ioutil.ReadFile("parole.txt")
	adjectives = strings.Split(string(foo), "\n")
}
