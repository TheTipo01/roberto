package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	tblCustomCommands = "CREATE TABLE IF NOT EXISTS \"customCommands\" (\"server\" VARCHAR(18) NOT NULL,\"command\" VARCHAR(50) NOT NULL,\"text\" VARCHAR(2000) NOT NULL);"
)

// findUserVoiceState finds the voicestate of a user
func findUserVoiceState(s *discordgo.Session, userID string, guidID string) *discordgo.VoiceState {
	for _, g := range s.State.Guilds {
		if g.ID == guidID {
			for _, vs := range g.VoiceStates {
				if vs.UserID == userID {
					return vs
				}
			}
		}
	}
	return nil
}

// advancedReplace returns src string with every instance of toReplace with a random item from a
func advancedReplace(src string, toReplace string, a []string) string {
	var dst = src

	for i := 0; i < strings.Count(src, toReplace); i++ {
		dst = strings.Replace(dst, toReplace, a[rand.Intn(len(a))], 1)
	}

	return dst
}

// Executes a simple query given a DB
func execQuery(query string, db *sql.DB) {
	stm, err := db.Prepare(query)
	if err != nil {
		lit.Error("Error preparing query, %s", err)
		return
	}

	_, err = stm.Exec()
	if err != nil {
		lit.Error("Error creating table, %s", err)
	}

	_ = stm.Close()
}

// Adds a custom command to db and to the command map
func addCommand(command string, text string, guild string) error {
	initializeServer(guild)

	// If the text is already in the map, we ignore it
	if server[guild].customCommands[command] == text {
		return errors.New("command already exists")
	}

	// Else, we add it to the map
	server[guild].customCommands[command] = text

	// And to the database
	stm, _ := db.Prepare("INSERT INTO customCommands (server, command, text) VALUES(?, ?, ?)")

	_, err := stm.Exec(guild, command, text)
	if err != nil {
		lit.Error("Error inserting into the database, %s", err)
		return errors.New("error inserting into the database: " + err.Error())
	}

	_ = stm.Close()

	return nil
}

// Removes a custom command from the db and from the command map
func removeCustom(command string, guild string) error {

	if server[guild].customCommands[command] == "" {
		return errors.New("command doesn't exist")
	}

	// Remove from DB
	stm, _ := db.Prepare("DELETE FROM customCommands WHERE server=? AND command=?")
	_, err := stm.Exec(guild, command)
	if err != nil {
		lit.Error("Error removing from the database, %s", err)
		return errors.New("error removing from the database: " + err.Error())
	}

	_ = stm.Close()

	// Remove from the map
	delete(server[guild].customCommands, command)

	return nil
}

// Loads custom command from the database
func loadCustomCommands(db *sql.DB) {
	var (
		guild, command, text string
		guilds, commands     *sql.Rows
		err                  error
	)

	guilds, err = db.Query("SELECT server FROM customCommands GROUP BY server")
	if err != nil {
		lit.Error("Error querying database, %s", err)
		return
	}

	for guilds.Next() {
		err = guilds.Scan(&guild)
		if err != nil {
			lit.Error("Error scanning server from query, %s", err)
			continue
		}

		initializeServer(guild)

		commands, err = db.Query("SELECT command, text FROM customCommands WHERE server=?", guild)
		if err != nil {
			lit.Error("Error querying database, %s", err)
			continue
		}

		for commands.Next() {
			err = commands.Scan(&command, &text)
			if err != nil {
				lit.Error("Error scanning commands from query, %s", err)
				continue
			}

			server[guild].customCommands[command] = text
		}

		_ = commands.Close()
	}

	_ = guilds.Close()
}

// Returns a random value from a map of string
func getRand(a map[string]string) string {
	// produce a pseudo-random number between 0 and len(a)-1
	i := int(float32(len(a)) * rand.Float32())
	for _, v := range a {
		if i == 0 {
			return v
		}
		i--
	}
	panic("impossible")
}

// Generates a bestemmia
func bestemmia() string {
	s1 := gods[rand.Intn(len(gods))]

	s := s1 + " " + adjectives[rand.Intn(len(adjectives))]

	if s1 == gods[2] {
		s = s[:len(s)-2] + "a"
	}

	return s
}

// Generates a DCA file starting from a string and it's UUID
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

// Initialize server for a given guildID if it's nil
func initializeServer(guildID string) {
	if server[guildID] == nil {
		server[guildID] = &Server{
			mutex:          &sync.Mutex{},
			stop:           true,
			customCommands: make(map[string]string),
		}
	}
}

// Sends embed as response to an interaction
func sendEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, c *chan int) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionApplicationCommandResponseData{Embeds: []*discordgo.MessageEmbed{embed}}})
	if err != nil {
		lit.Error("InteractionRespond failed: %s", err)
		return
	}

	if c != nil {
		*c <- 1
	}
}

// Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration) {
	sendEmbedInteraction(s, embed, i, nil)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(s.State.User.ID, i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// Modify an already sent interaction
func modfyInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, c *chan int) {
	err := s.InteractionResponseEdit(s.State.User.ID, i, &discordgo.WebhookEdit{Embeds: []*discordgo.MessageEmbed{embed}})
	if err != nil {
		lit.Error("InteractionResponseEdit failed: %s", err)
		return
	}

	if c != nil {
		*c <- 1
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
