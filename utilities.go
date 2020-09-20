package main

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strings"
)

const (
	tblCustomCommands = "CREATE TABLE IF NOT EXISTS \"customCommands\" (\"server\" VARCHAR(18) NOT NULL,\"command\" VARCHAR(50) NOT NULL,\"text\" VARCHAR(2000) NOT NULL);"
)

// deleteMessage delete a message
func deleteMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println(m.Author.Username + ": " + m.Content)
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		fmt.Println("Can't delete message,", err)
	}
}

// findUserVoiceState finds the voicestate of a user
func findUserVoiceState(session *discordgo.Session, userid string) *discordgo.VoiceState {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs
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
	statement, err := db.Prepare(query)
	if err != nil {
		log.Println("Error preparing query,", err)
		return
	}

	_, err = statement.Exec()
	if err != nil {
		log.Println("Error creating table,", err)
	}
}

// Adds a custom command to db and to the command map
func addCommand(command string, text string, guild string) {
	// If the text is already in the map, we ignore it
	if customCommands[guild][command] == text {
		return
	}

	if customCommands[guild] == nil {
		customCommands[guild] = make(map[string]string)
	}

	// Else, we add it to the map
	customCommands[guild][command] = text

	// And to the database
	statement, _ := db.Prepare("INSERT INTO customCommands (server, command, text) VALUES(?, ?, ?)")

	_, err := statement.Exec(guild, command, text)
	if err != nil {
		log.Println("Error inserting into the database,", err)
	}

}

// Removes a custom command from the db and from the command map
func removeCustom(command string, guild string) {
	// Remove from DB
	statement, _ := db.Prepare("DELETE FROM customCommands WHERE server=? AND command=?")
	_, err := statement.Exec(guild, command)
	if err != nil {
		log.Println("Error removing from the database,", err)
	}

	// Remove from the map
	delete(customCommands[guild], command)
}

// Loads custom command from the database
func loadCustomCommands(db *sql.DB) {
	var guild, command, text string

	rows, err := db.Query("SELECT * FROM customCommands")
	if err != nil {
		log.Println("Error querying database,", err)
	}

	for rows.Next() {
		err = rows.Scan(&guild, &command, &text)
		if err != nil {
			log.Println("Error scanning rows from query,", err)
			continue
		}

		if customCommands[guild] == nil {
			customCommands[guild] = make(map[string]string)
		}

		customCommands[guild][command] = text
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
