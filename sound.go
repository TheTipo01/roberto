package main

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/TheTipo01/roberto/queue"
)

// Plays a song in DCA format
func playSound(guildID string, el *queue.Element) bool {
	var (
		opusLen int16
		err     error
	)

	if server[guildID].vc == nil {
		return false
	}
	_ = server[guildID].vc.Speaking(true)

	for {
		select {
		case <-server[guildID].skip:
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

			server[guildID].vc.OpusSend <- InBuf
		}
	}
}

func cleanUp(guildID string, closer io.Closer) {
	_ = server[guildID].vc.Speaking(false)

	if closer != nil {
		_ = closer.Close()
	}
}
