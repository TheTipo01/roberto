package main

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
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
	token  string
	server = make(map[string]*sync.Mutex)
	stop   = make(map[string]bool)
	b      []string
	a      = [3]string{"Dio", "Gesù", "Madonna"}
)

func bestemmia() string {

	s1 := a[rand.Intn(len(a))]

	s := s1 + " " + b[rand.Intn(len(b))]

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

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	viper.SetDefault("prefix", "!")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			fmt.Println("Config file not found! See example_config.yml")
			return
		}
	} else {
		//Config file found
		token = viper.GetString("token")
	}

	//Read adjective
	foo, _ := ioutil.ReadFile("parole.txt")
	b = strings.Split(string(foo), "\n")

	//Initialize rand
	rand.Seed(time.Now().Unix())
}

func main() {

	if token == "" {
		fmt.Println("No token provided. Please run: roberto -token <bot token> or modify config.yml")
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

	//We set the intents that we use
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
	err := s.UpdateStatus(0, "!say, !covid, !bestemmia 1")
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
	//Makes the message all uppercase and replaces endlines with blank spaces
	message := strings.ReplaceAll(strings.ToLower(m.Content), "\n", " ")

	// check if the message is "!bestemmia"
	if strings.HasPrefix(message, "!bestemmia") {
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs == nil {
			return
		}

		splitted := strings.Split(m.Content, " ")

		//Locks the mutex for the current server
		server[vs.GuildID].Lock()

		// Join the provided voice channel.
		vc, err := s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
		if err != nil {
			fmt.Println("Can't connect to voice channel,", err)
			server[vs.GuildID].Unlock()
			return
		}

		//If a number possibly exist
		if len(splitted) > 1 {
			n, err := strconv.Atoi(splitted[1])
			if err == nil {
				//And we can convert it to a n, we repeat the sound for n times

				for i := 0; i < n; i++ {
					if stop[vs.GuildID] {
						playSound2(genAudio(strings.ToUpper(bestemmia())), vc)
					} else {
						//Resets the stop boolean
						stop[vs.GuildID] = true
						break
					}
				}

			}
		} else {
			//Else, we only do the command once
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

		return
	}

	if strings.HasPrefix(message, "!say") {
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(strings.TrimPrefix(message, "!say ")))
		}

		return
	}

	if strings.HasPrefix(message, "!stop") {
		stop[m.GuildID] = false
		go deleteMessage(s, m)
		return
	}

	if strings.HasPrefix(message, "!treno") {
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(ricercaAndGetTreno(strings.TrimPrefix(message, "!treno "))))
		}

		return
	}

	if strings.HasPrefix(message, "!covid") {
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m.Author.ID)
		if vs != nil {
			playSound(s, vs.GuildID, vs.ChannelID, genAudio(getCovid()))
		}

		return
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

	//Locks the mutex for the current server
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

	//Resets the stop boolean
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
