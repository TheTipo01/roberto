package main

import (
	"github.com/bwmarrin/lit"
	"github.com/forPelevin/gomoji"
	"github.com/goccy/go-json"
	"os"
	"strings"
)

func emojiReplacer() *strings.Replacer {
	var (
		emojiJSON Emoji
		args      []string
	)

	// Load JSON file
	jsonFile, err := os.Open("emoji.json")
	if err != nil {
		lit.Error("Error opening file: %s", err)
		return nil
	}

	_ = json.NewDecoder(jsonFile).Decode(&emojiJSON)
	_ = jsonFile.Close()

	// Create the replacer
	for _, e := range emojiJSON {
		args = append(args, e.Emoji, e.Descrizione)
	}

	return strings.NewReplacer(args...)
}

func emojiToDescription(str string) string {
	if gomoji.ContainsEmoji(str) {
		str = emoji.Replace(str)
	}

	return strings.ToLower(str)
}
