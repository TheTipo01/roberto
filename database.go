package main

import (
	"database/sql"
	"errors"
	"github.com/bwmarrin/lit"
)

// Table creation queries
const (
	tblCustomCommands = "CREATE TABLE IF NOT EXISTS \"customCommands\" (\"server\" VARCHAR(18) NOT NULL,\"command\" VARCHAR(50) NOT NULL,\"text\" VARCHAR(2000) NOT NULL);"
)

// DB parameters
const (
	dataSourceName = "./roberto.db"
	driverName     = "sqlite"
)

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
