package main

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/goccy/go-json"
	"math/rand"
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
	_, err := db.Exec(query)
	if err != nil {
		lit.Error("Error preparing query, %s", err)
		return
	}
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
	_, err := db.Exec("INSERT INTO customCommands (server, command, text) VALUES(?, ?, ?)", guild, command, text)
	if err != nil {
		lit.Error("Error inserting into the database, %s", err)
		return errors.New("error inserting into the database: " + err.Error())
	}

	return nil
}

// Removes a custom command from the db and from the command map
func removeCustom(command string, guild string) error {

	if server[guild].customCommands[command] == "" {
		return errors.New("command doesn't exist")
	}

	// Remove from DB
	_, err := db.Exec("DELETE FROM customCommands WHERE server=? AND command=?", guild, command)
	if err != nil {
		lit.Error("Error removing from the database, %s", err)
		return errors.New("error removing from the database: " + err.Error())
	}

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
	}
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
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}}})
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
	_, err := s.InteractionResponseEdit(s.State.User.ID, i, &discordgo.WebhookEdit{Embeds: []*discordgo.MessageEmbed{embed}})
	if err != nil {
		lit.Error("InteractionResponseEdit failed: %s", err)
		return
	}

	if c != nil {
		*c <- 1
	}
}

// isCommandEqual compares two command by marshalling them to JSON. Yes, I know. I don't want to write recursive things.
func isCommandEqual(c *discordgo.ApplicationCommand, v *discordgo.ApplicationCommand) bool {
	c.Version = ""
	c.ID = ""
	c.ApplicationID = ""
	c.Type = 0
	cBytes, _ := json.Marshal(&c)

	v.Version = ""
	v.ID = ""
	v.ApplicationID = ""
	v.Type = 0
	vBytes, _ := json.Marshal(&v)

	return bytes.Compare(cBytes, vBytes) == 0
}
