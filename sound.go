package main

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/TheTipo01/roberto/queue"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

// Plays a song in DCA format
func playSound(guildID snowflake.ID, el *queue.Element) bool {
	var (
		opusLen int16
		err     error
	)

	srv := getServer(guildID)

	if srv.vc == nil {
		return false
	}
	_ = srv.vc.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone)

	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		select {
		case <-srv.skip:
			cleanUp(guildID, el.Closer)
			return true
		default:
			// Read opus frame length from dca file.
			err = binary.Read(el.Reader, binary.LittleEndian, &opusLen)

			// If this is the end of the file, just return.
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				cleanUp(guildID, el.Closer)
				return false
			}

			// Read encoded pcm from dca file.
			InBuf := make([]byte, opusLen)
			err = binary.Read(el.Reader, binary.LittleEndian, &InBuf)

			// Should not be any end of file errors
			if err != nil {
				cleanUp(guildID, el.Closer)
				return false
			}

			_, err = srv.vc.UDP().Write(InBuf)
			if err != nil {
				cleanUp(guildID, el.Closer)
				return false
			}
		}
	}

	return true
}

func cleanUp(guildID snowflake.ID, closer io.Closer) {
	srv := getServer(guildID)

	_ = srv.vc.SetSpeaking(context.TODO(), voice.SpeakingFlagNone)

	if closer != nil {
		_ = closer.Close()
	}
}
