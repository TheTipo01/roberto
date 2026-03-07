package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	libroberto "github.com/TheTipo01/libRoberto"
	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/godave/golibdave"
	"github.com/disgoorg/snowflake/v2"
	"github.com/kkyr/fig"
	_ "modernc.org/sqlite"
)

type config struct {
	Token            string `fig:"token" validate:"required"`
	LogLevel         string `fig:"loglevel" validate:"required"`
	Voice            string `fig:"voice" validate:"required"`
	RestRoberto      string `fig:"restroberto"`
	RestRobertoToken string `fig:"restrobertotoken"`
}

var (
	// Discord bot token
	token string
	// Server
	server = make(map[snowflake.ID]*Server)
	// DB connection
	db *sql.DB
	// Discord bot session
	s *bot.Client
	// Endpoint for rest roberto
	restRoberto string
	// Token for rest roberto
	restRobertoToken string
	// Channel used to notify the presence updater that the guild count has changed
	guildCountChan = make(chan struct{})
	// BotName is the name of the bot
	BotName string
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
	restRoberto = cfg.RestRoberto
	restRobertoToken = cfg.RestRobertoToken

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

	go presenceUpdater()
}

func main() {
	if token == "" {
		lit.Error("No token provided. Please modify config.yml")
		return
	}

	logger := slog.Default()
	if lit.LogLevel == lit.LogDebug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildVoiceStates,
				gateway.IntentGuilds,
			),
		),

		bot.WithCacheConfigOpts(
			cache.WithCaches(
				cache.FlagVoiceStates,
			),
		),

		bot.WithEventListenerFunc(ready),
		bot.WithEventListenerFunc(guildCreate),
		bot.WithEventListenerFunc(guildDelete),
		bot.WithEventListenerFunc(interactionCreate),

		bot.WithVoiceManagerConfigOpts(voice.WithDaveSessionCreateFunc(golibdave.NewSession)),

		bot.WithLogger(logger),
	)
	
	if err != nil {
		lit.Error("Error creating bot client: %s", err)
		return
	}

	defer client.Close(context.TODO())

	if err := client.OpenGateway(context.TODO()); err != nil {
		lit.Error("errors while connecting to gateway %s", err)
		return
	}

	// Save the session
	s = client

	// Register commands
	_, err = client.Rest.SetGlobalCommands(client.ApplicationID, commands)
	if err != nil {
		lit.Error("Error registering commands: %s", err)
		return
	}

	// Wait here until CTRL-C or another term signal is received.
	lit.Info("roberto is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	_ = db.Close()
}

// presenceUpdater updates the bot presence every time the guild count changes, with a debounce of 500ms to avoid making too many requests
func presenceUpdater() {
	debounceTimer := time.NewTimer(0)
	debounceTimer.Stop()

	for {
		select {
		case <-guildCountChan:
			debounceTimer.Reset(500 * time.Millisecond)
		case <-debounceTimer.C:
			if s != nil {
				_ = s.SetPresence(context.TODO(), gateway.WithCustomActivity("Serving "+strconv.Itoa(len(server))+" guilds!"))
			}
		}
	}
}

func notifyGuildCountChange() {
	select {
	case guildCountChan <- struct{}{}:
	default:
	}
}

func ready(e *events.Ready) {
	notifyGuildCountChange()

	BotName = e.User.Username
}

func guildCreate(e *events.GuildReady) {
	initializeServer(e.GuildID)
	notifyGuildCountChange()
}

func guildDelete(e *events.GuildLeave) {
	notifyGuildCountChange()
}

func interactionCreate(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	// Ignores commands from DM
	if e.Context() == discord.InteractionContextTypeGuild {
		if h, ok := commandHandlers[data.CommandName()]; ok {
			go h(e)
		}
	} else {
		go sendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField("Error",
			"Commands are not available in DM!", false).
			WithColor(0x7289DA), e, time.Second*15)
	}
}
