package main

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/disgoorg/snowflake/v2"
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
		slog.Error("error preparing query", "error", err)
		return
	}
}

// Adds a custom command to db and to the command map
func addCommand(command string, text string, guild snowflake.ID) error {
	srv := getServer(guild)

	// If the text is already in the map, we ignore it
	if existing, ok := srv.GetCustomCommand(command); ok && existing == text {
		return errors.New("command already exists")
	}

	// Else, we add it to the map
	srv.SetCustomCommand(command, text)

	// And to the database
	_, err := db.Exec("INSERT INTO customCommands (server, command, text) VALUES(?, ?, ?)", guild, command, text)
	if err != nil {
		slog.Error("error inserting into the database", "error", err)
		return errors.New("error inserting into the database: " + err.Error())
	}

	return nil
}

// Removes a custom command from the db and from the command map
func removeCustom(command string, guild snowflake.ID) error {
	srv := getServer(guild)

	if _, ok := srv.GetCustomCommand(command); !ok {
		return errors.New("command doesn't exist")
	}

	// Remove from DB
	_, err := db.Exec("DELETE FROM customCommands WHERE server=? AND command=?", guild, command)
	if err != nil {
		slog.Error("error removing from the database", "error", err)
		return errors.New("error removing from the database: " + err.Error())
	}

	// Remove from the map
	srv.DeleteCustomCommand(command)

	return nil
}

// Loads custom command from the database
func loadCustomCommands(db *sql.DB) {
	var (
		guild, command, text string
		guildSnowflake       snowflake.ID
		guilds, commands     *sql.Rows
		err                  error
	)

	guilds, err = db.Query("SELECT server FROM customCommands GROUP BY server")
	if err != nil {
		slog.Error("error querying database", "error", err)
		return
	}

	for guilds.Next() {
		err = guilds.Scan(&guild)
		if err != nil {
			slog.Error("error scanning server from query", "error", err)
			continue
		}

		guildSnowflake = snowflake.MustParse(guild)

		srv := getServer(guildSnowflake)

		commands, err = db.Query("SELECT command, text FROM customCommands WHERE server=?", guild)
		if err != nil {
			slog.Error("error querying database", "error", err)
			continue
		}

		for commands.Next() {
			err = commands.Scan(&command, &text)
			if err != nil {
				slog.Error("error scanning commands from query", "error", err)
				continue
			}

			srv.SetCustomCommand(command, text)
		}
	}
}
