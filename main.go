package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	// Discord bot token
	token string
	// Prefix for discord commands
	prefix string
	// Mutex for syncing requests
	server = make(map[string]*sync.Mutex)
	// Boolean for skipping
	stop = make(map[string]bool)
	// Custom commands
	customCommands = make(map[string]map[string]string)
	// Array of adjectives
	adjectives []string
	// Gods
	gods = []string{"Dio", "Gesù", "Madonna"}
	// DB Stuff
	dataSourceName = "./roberto.db"
	driverName     = "sqlite3"
	db             *sql.DB
)

func bestemmia() string {

	s1 := gods[rand.Intn(len(gods))]

	s := s1 + " " + adjectives[rand.Intn(len(adjectives))]

	if s1 == "Madonna" {
		s = s[:len(s)-2] + "a"
	}

	return s
}

func gen(bestemmia string, uuid string) {
	_, err := os.Stat("./temp/" + uuid + ".dca")

	if err != nil {
		switch runtime.GOOS {
		case "linux":
			cmd := exec.Command("/bin/bash", "gen.sh", uuid, bestemmia)
			_ = cmd.Run()
		case "windows":
			cmd := exec.Command("gen.bat", uuid)
			cmd.Stdin = strings.NewReader(bestemmia)
			_ = cmd.Run()
		}

	}

}

func init() {
	var err error

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	viper.SetDefault("prefix", "!")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			fmt.Println("Config file not found! See example_config.yml")
			return
		}
	} else {
		// Config file found
		token = viper.GetString("token")
		prefix = viper.GetString("prefix")
	}

	// Read adjective
	foo, _ := ioutil.ReadFile("parole.txt")
	adjectives = strings.Split(string(foo), "\n")

	// Initialize rand
	rand.Seed(time.Now().Unix())

	// Database
	db, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Println("Error opening db connection,", err)
		return
	}

	execQuery(tblCustomCommands, db)

	loadCustomCommands(db)
}

func main() {

	if token == "" {
		fmt.Println("No token provided. Please modify config.yml")
		return
	}

	if prefix == "" {
		fmt.Println("No prefix provided. Please modify config.yml")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
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
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("roberto is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()
}

func ready(s *discordgo.Session, _ *discordgo.Ready) {

	// Set the playing status.
	err := s.UpdateStatus(0, prefix+"help")
	if err != nil {
		fmt.Println("Can't set status,", err)
	}
}

func guildCreate(_ *discordgo.Session, event *discordgo.GuildCreate) {
	server[event.ID] = &sync.Mutex{}
	stop[event.ID] = true
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Makes the lowerMessage all uppercase and replaces endlines with blank spaces
	lowerMessage := strings.ToLower(m.Content)

	switch strings.Split(lowerMessage, " ")[0] {

	case prefix + "bestemmia":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs == nil {
			return
		}

		splitted := strings.Split(m.Content, " ")

		// Locks the mutex for the current server
		server[vs.GuildID].Lock()

		// Join the provided voice channel.
		vc, err := s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
		if err != nil {
			fmt.Println("Can't connect to voice channel,", err)
			server[vs.GuildID].Unlock()
			return
		}

		// If a number possibly exist
		if len(splitted) > 1 {
			n, err := strconv.Atoi(splitted[1])
			if err == nil {
				// And we can convert it to a n, we repeat the sound for n times

				for i := 0; i < n; i++ {
					if stop[vs.GuildID] {
						playSound2(genAudio(strings.ToUpper(bestemmia())), vc)
					} else {
						// Resets the stop boolean
						stop[vs.GuildID] = true
						break
					}
				}

			}
		} else {
			// Else, we only do the command once
			playSound2(genAudio(strings.ToUpper(bestemmia())), vc)
		}

		// Disconnect from the provided voice channel.
		err = vc.Disconnect()
		if err != nil {
			fmt.Println("Can't disconnect from voice channel,", err)
			server[vs.GuildID].Unlock()
			return
		}

		// Releases the mutex lock for the server
		server[vs.GuildID].Unlock()

		break

	case prefix + "say":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(strings.TrimPrefix(lowerMessage, prefix+"say ")))
		}

		break

	case prefix + "stop":
		stop[m.GuildID] = false
		go deleteMessage(s, m)
		break

	case prefix + "treno":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(ricercaAndGetTreno(strings.TrimPrefix(lowerMessage, prefix+"treno "))))
		}

		break

	case prefix + "covid":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(getCovid()))
		}

		break

		// Adds a custom command
	case prefix + "custom":
		go deleteMessage(s, m)

		splitted := strings.Split(lowerMessage, " ")

		if len(splitted) > 2 {
			addCommand(splitted[1], strings.TrimPrefix(lowerMessage, prefix+"custom "+splitted[1]+" "), m.GuildID)
		}
		break

		// Removes a custom command
	case prefix + "rmcustom":
		go deleteMessage(s, m)

		removeCustom(strings.TrimPrefix(m.Content, prefix+"rmcustom "), m.GuildID)
		break

	case prefix + "preghiera":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(advancedReplace(advancedReplace(GetRand(customCommands[m.GuildID]), "<god>", gods), "<dict>", adjectives)))
		}

		break

		// Prints out supported commands
	case prefix + "help", prefix + "h":
		go deleteMessage(s, m)

		message := "Supported commands:\n```" +
			prefix + "say <text> - Says text out loud\n" +
			prefix + "bestemmioa <n> - Generates a bestemmia n times\n" +
			prefix + "treno <train number> - Fakes train announcement given it's number\n" +
			prefix + "covid - Says covid data out loud for current day in Italy\n" +
			prefix + "preghiera - Randomly select a custom command\n" +
			prefix + "custom <custom command> <text> - Creates a custom command to say text out loud. The bot will replace <god> with a random god and <dict> with a random adjective\n" +
			prefix + "rmcustom <custom command> - Removes a custom command\n" +
			"```"
		// If we have custom commands, we add them to the help message
		if len(customCommands[m.GuildID]) > 0 {
			message += "\nCustom commands:\n```"

			for k := range customCommands[m.GuildID] {
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
		lower := strings.TrimPrefix(lowerMessage, prefix)

		if customCommands[m.GuildID][lower] != "" {
			go deleteMessage(s, m)

			vs := findUserVoiceState(s, m.Author.ID)
			if vs != nil {
				playSound(s, vs.GuildID, vs.ChannelID, genAudio(advancedReplace(advancedReplace(customCommands[m.GuildID][lower], "<god>", gods), "<dict>", adjectives)))
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
		fmt.Println("Error opening dca file :", err)
		return
	}
	defer file.Close()

	// Locks the mutex for the current server
	server[guildID].Lock()

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return
	}

	// Start speaking.
	_ = vc.Speaking(true)
	stop[guildID] = true

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			break
		}

		// Stream data to discord
		if stop[guildID] {
			vc.OpusSend <- InBuf
		} else {
			break
		}
	}

	// Resets the stop boolean
	stop[guildID] = true

	// Stop speaking
	_ = vc.Speaking(false)

	// Disconnect from the provided voice channel.
	err = vc.Disconnect()
	if err != nil {
		fmt.Println("Can't disconnect from voice channel,", err)
		return
	}

	// Releases the mutex lock for the server
	server[guildID].Unlock()

}

// playSound2 plays a file to the provided channel given a voice connection.
func playSound2(fileName string, vc *discordgo.VoiceConnection) {
	var opuslen int16

	file, err := os.Open("./temp/" + fileName)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return
	}
	defer file.Close()

	// Start speaking.
	_ = vc.Speaking(true)

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			break
		}

		// Stream data to discord
		vc.OpusSend <- InBuf

	}

	// Stop speaking
	_ = vc.Speaking(false)

}
