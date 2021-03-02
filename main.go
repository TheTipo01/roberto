package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/spf13/viper"
	"io"
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
	// Prefix for discord commands
	prefix string
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
		prefix = viper.GetString("prefix")

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

		// Read adjective
		foo, _ := ioutil.ReadFile("parole.txt")
		adjectives = strings.Split(string(foo), "\n")

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

	if prefix == "" {
		lit.Error("No prefix provided. Please modify config.yml")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		lit.Error("Error creating Discord session: %s", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(ready)

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
	err := s.UpdateGameStatus(0, prefix+"help")
	if err != nil {
		lit.Error("Can't set status, %s", err)
	}
}

func guildCreate(_ *discordgo.Session, e *discordgo.GuildCreate) {
	initializeServer(e.Guild.ID)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages sent from the bot, messages if the user is a bot, and messages without the prefix
	if s.State.User.ID == m.Author.ID || m.Author.Bot || !strings.HasPrefix(m.Content, prefix) {
		return
	}

	// Split the message on spaces
	splittedMessage := strings.Split(m.Content, " ")

	command := strings.TrimPrefix(strings.ToLower(splittedMessage[0]), prefix)

	lowerMessage := strings.ToLower(strings.TrimPrefix(m.Content, splittedMessage[0]))

	switch command {
	case "bestemmia":
		go deleteMessage(s, m.Message)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs == nil {
			return
		}

		// Locks the mutex for the current server
		server[vs.GuildID].mutex.Lock()

		// Join the provided voice channel.
		vc, err := s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
		if err != nil {
			lit.Error("Can't connect to voice channel, %s", err)
			server[vs.GuildID].mutex.Unlock()
			return
		}

		// If a number possibly exist
		if len(splittedMessage) > 1 {
			n, err := strconv.Atoi(splittedMessage[1])
			if err == nil {
				// And we can convert it to a n, we repeat the sound for n times

				for i := 0; i < n; i++ {
					if server[vs.GuildID].stop {
						playSound2(genAudio(strings.ToUpper(bestemmia())), vc, s)
					} else {
						// Resets the stop boolean
						server[vs.GuildID].stop = true
						break
					}
				}
			}
		} else {
			// Else, we only do the command once
			playSound2(genAudio(strings.ToUpper(bestemmia())), vc, s)
		}

		// Disconnect from the provided voice channel.
		err = vc.Disconnect()
		if err != nil {
			lit.Error("Can't disconnect from voice channel, %s", err)
		}

		// Releases the mutex lock for the server
		server[vs.GuildID].mutex.Unlock()
		break

	case "say":
		go deleteMessage(s, m.Message)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(emojiToDescription(strings.TrimPrefix(lowerMessage, prefix+"say "))))
		}
		break

	case "stop":
		server[m.GuildID].stop = false
		go deleteMessage(s, m.Message)
		break

	case "treno":
		go deleteMessage(s, m.Message)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(searchAndGetTrain(strings.TrimPrefix(lowerMessage, prefix+"treno "))))
		}
		break

	case "covid":
		go deleteMessage(s, m.Message)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			covid := getCovid()

			_, _ = s.ChannelMessageSend(m.ChannelID, covid)
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(covid))
		}
		break

		// Adds a custom command
	case "custom":
		go deleteMessage(s, m.Message)

		if len(splittedMessage) > 2 {
			addCommand(splittedMessage[1], strings.TrimPrefix(lowerMessage, prefix+"custom "+splittedMessage[1]+" "), m.GuildID)
		}
		break

		// Removes a custom command
	case "rmcustom":
		go deleteMessage(s, m.Message)

		removeCustom(strings.TrimPrefix(m.Content, prefix+"rmcustom "), m.GuildID)
		break

	case prefix + "preghiera":
		go deleteMessage(s, m.Message)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(advancedReplace(advancedReplace(getRand(server[m.GuildID].customCommands), "<god>", gods), "<dict>", adjectives)))
		}
		break

		// Prints out supported commands
	case "help", "h":
		go deleteMessage(s, m.Message)

		message := "Supported commands:\n```" +
			prefix + "say <text> - Says text out loud\n" +
			prefix + "bestemmia <n> - Generates a bestemmia n times\n" +
			prefix + "treno <train number> - Fakes train announcement given it's number\n" +
			prefix + "covid - Says covid data out loud for current day in Italy\n" +
			prefix + "preghiera - Randomly select a custom command\n" +
			prefix + "custom <custom command> <text> - Creates a custom command to say text out loud. The bot will replace <god> with a random god and <dict> with a random adjective\n" +
			prefix + "rmcustom <custom command> - Removes a custom command\n" +
			"```"
		// If we have custom commands, we add them to the help message
		if len(server[m.GuildID].customCommands) > 0 {
			message += "\nCustom commands:\n```"

			for k := range server[m.GuildID].customCommands {
				message += k + ", "
			}

			message = strings.TrimSuffix(message, ", ")
			message += "```"
		}

		mex, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			fmt.Println(err)
			break
		}

		time.Sleep(time.Second * 30)

		err = s.ChannelMessageDelete(m.ChannelID, mex.ID)
		if err != nil {
			fmt.Println(err)
		}
		break

		// We search for possible custom commands
	default:
		if server[m.GuildID].customCommands[command] != "" {
			go deleteMessage(s, m.Message)

			vs := findUserVoiceState(s, m.Author.ID)
			if vs != nil {
				playSound(s, vs.GuildID, vs.ChannelID, genAudio(advancedReplace(advancedReplace(server[m.GuildID].customCommands[command], "<god>", gods), "<dict>", adjectives)))
			}
			break
		}
	}
}

// genAudio generates a dca file from a string
func genAudio(text string) string {
	h := sha1.New()
	h.Write([]byte(text))
	uuid := strings.ToUpper(base32.HexEncoding.EncodeToString(h.Sum(nil)))

	gen(text, uuid)

	return uuid + ".dca"
}

// playSound plays a file to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID, fileName string) {
	var opuslen int16

	file, err := os.Open("./temp/" + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}
	defer file.Close()

	// Locks the mutex for the current server
	server[guildID].mutex.Lock()

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		server[guildID].mutex.Unlock()
		return
	}

	// Start speaking.
	_ = vc.Speaking(true)
	server[guildID].stop = true

	// Channel to send ok messages
	c1 := make(chan string, 1)

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Stream data to discord
		if server[guildID].stop {
			// Send data in a goroutine
			go func() {
				vc.OpusSend <- InBuf
				c1 <- "ok"
			}()

			// So if the bot gets disconnect/moved we can rejoin the original channel and continue playing songs
			select {
			case _ = <-c1:
				break
			case <-time.After(time.Second / 3):
				vc, _ = s.ChannelVoiceJoin(guildID, channelID, false, true)
			}
		} else {
			break
		}
	}

	// Resets the stop boolean
	server[guildID].stop = true

	// Stop speaking
	_ = vc.Speaking(false)

	// Disconnect from the provided voice channel.
	err = vc.Disconnect()
	if err != nil {
		lit.Error("Can't disconnect from voice channel, %s", err)
		return
	}

	// Releases the mutex lock for the server
	server[guildID].mutex.Unlock()
}

// playSound2 plays a file to the provided channel given a voice connection.
func playSound2(fileName string, vc *discordgo.VoiceConnection, s *discordgo.Session) {
	var opuslen int16

	file, err := os.Open("./temp/" + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}
	defer file.Close()

	// Start speaking.
	_ = vc.Speaking(true)

	// Channel to send ok messages
	c1 := make(chan string, 1)

	guildID := vc.GuildID
	channelID := vc.ChannelID

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Stream data to discord
		// Send data in a goroutine
		go func() {
			vc.OpusSend <- InBuf
			c1 <- "ok"
		}()

		// So if the bot gets disconnect/moved we can rejoin the original channel and continue playing songs
		select {
		case _ = <-c1:
			break
		case <-time.After(time.Second / 3):
			vc, _ = s.ChannelVoiceJoin(guildID, channelID, false, true)
		}

	}

	// Stop speaking
	_ = vc.Speaking(false)
}
