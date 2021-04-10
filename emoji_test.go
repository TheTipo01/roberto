package main

import (
	"strings"
	"testing"
)

func TestEmojiReplacer(t *testing.T) {
	emoji = *emojiReplacer()

	if &emoji == nil {
		t.Error("Emoji replacer is empty")
	}
}

func TestEmojiToDescription(t *testing.T) {
	if &emoji == nil {
		emoji = *emojiReplacer()
	}

	if strings.ToLower(emoji.Replace("🕴️")) != "uomo con completo che levita" {
		t.Error("Emoji replacer gave wrong description")
	}
}
