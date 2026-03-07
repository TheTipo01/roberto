package main

import (
	"sync/atomic"

	"github.com/TheTipo01/roberto/queue"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// NewServer creates a new server manager
func NewServer(guildID snowflake.ID) *Server {
	return &Server{
		skip:           make(chan struct{}),
		guildID:        guildID,
		customCommands: make(map[string]string),
		queue:          queue.NewQueue(),
		started:        atomic.Bool{},
		clear:          atomic.Bool{},
	}
}

// AddSong adds a song to the queue
func (m *Server) AddSong(priority bool, el ...queue.Element) {
	if priority {
		m.queue.AddElementsPriority(el...)
	} else {
		m.queue.AddElements(el...)
	}

	if m.started.CompareAndSwap(false, true) {
		go m.play()
	}
}

func (m *Server) play() {
	msg := make(chan *discord.Message)

	for el := m.queue.GetFirstElement(); el != nil && !m.clear.Load(); el = m.queue.GetFirstElement() {
		// Send "Now playing" message
		go func() {
			msg <- sendEmbed(s, discord.NewEmbedBuilder().SetTitle(BotName).
				AddField(el.Type, el.Content, false).
				SetColor(0x7289DA).Build(), el.TextChannel)
		}()

		if el.BeforePlay != nil {
			el.BeforePlay()
		}

		playSound(m.guildID, el)

		if el.AfterPlay != nil {
			el.AfterPlay()
		}

		// Delete it after it has been played
		go func() {
			if message := <-msg; message != nil {
				_ = s.Rest.DeleteMessage(message.ChannelID, message.ID)
			}
		}()

		m.queue.RemoveFirstElement()
	}

	m.started.Store(false)

	go quitVC(m.guildID)
}

// IsPlaying returns whether the bot is playing
func (m *Server) IsPlaying() bool {
	return m.started.Load() && !m.queue.IsEmpty()
}

// Clear clears the queue
func (m *Server) Clear() {
	if m.IsPlaying() {
		m.clear.Store(true)
		m.skip <- struct{}{}

		m.clear.Store(false)

		q := m.queue.GetAllQueue()
		m.queue.Clear()

		for _, el := range q {
			if el.Closer != nil {
				_ = el.Closer.Close()
			}
		}
	}
}
