package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	libroberto "github.com/TheTipo01/libRoberto"
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
	// Mutex guarding the server map
	serverMutex sync.RWMutex
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

// parseLogLevel maps a config loglevel value to a slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "logwarning", "warning":
		return slog.LevelWarn

	case "loginformational", "informational":
		return slog.LevelInfo

	case "logdebug", "debug":
		return slog.LevelDebug

	default:
		return slog.LevelError
	}
}

// setLogger configures the default logger, writing to stdout at the given level
func setLogger(logLevel string) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(logLevel)})))
}

func init() {
	// Log only errors until the configuration is loaded
	setLogger("error")

	var cfg config
	err := fig.Load(&cfg, fig.File("config.yml"))
	if err != nil {
		slog.Error("error loading config", "error", err)
		return
	}

	setLogger(cfg.LogLevel)

	libroberto.Voice = cfg.Voice
	token = cfg.Token
	restRoberto = cfg.RestRoberto
	restRobertoToken = cfg.RestRobertoToken

	// Database
	db, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		slog.Error("error opening database connection", "error", err)
		return
	}

	execQuery(tblCustomCommands, db)

	loadCustomCommands(db)

	go presenceUpdater()
}

func main() {
	if token == "" {
		slog.Error("no token provided, please modify config.yml")
		return
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
		bot.WithEventListenerFunc(guildJoin),
		bot.WithEventListenerFunc(guildDelete),
		bot.WithEventListenerFunc(interactionCreate),

		bot.WithVoiceManagerConfigOpts(voice.WithDaveSessionCreateFunc(golibdave.NewSession)),

		bot.WithLogger(slog.Default()),
	)

	if err != nil {
		slog.Error("error creating bot client", "error", err)
		return
	}

	defer client.Close(context.TODO())

	if err := client.OpenGateway(context.TODO()); err != nil {
		slog.Error("errors while connecting to gateway", "error", err)
		return
	}

	// Save the session
	s = client

	// Register commands
	_, err = client.Rest.SetGlobalCommands(client.ApplicationID, commands)
	if err != nil {
		slog.Error("error registering commands", "error", err)
		return
	}

	// Wait here until CTRL-C or another term signal is received.
	slog.Info("roberto is now running. Press CTRL-C to exit.")
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
				_ = s.SetPresence(context.TODO(), gateway.WithCustomActivity("Serving "+strconv.Itoa(guildCount())+" guilds!"))
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

// guildJoin is called when the bot is added to a guild while it's running.
// Without this, the guild would never be added to the server map and any
// command would crash with a nil pointer dereference.
func guildJoin(e *events.GuildJoin) {
	initializeServer(e.GuildID)
	notifyGuildCountChange()
}

func guildDelete(_ *events.GuildLeave) {
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
