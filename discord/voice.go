package discord

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type VoiceSession struct {
	connection    *discordgo.VoiceConnection
	buffer        [][]byte
	bufferUpdated chan struct{}
	quit          chan struct{}
	speaking      bool
	bot           *Bot
}

func NewVoice(s *discordgo.Session, bot *Bot, GuildID string, ChannelID string) (*VoiceSession, error) {

	if bot.stopping {
		return nil, errors.New("bot stopping")
	}

	var err error
	ret := &VoiceSession{
		buffer:        make([][]byte, 0),
		bufferUpdated: make(chan struct{}),
		quit:          make(chan struct{}),
		bot:           bot,
	}

	ret.connection, err = s.ChannelVoiceJoin(GuildID, ChannelID, false, true)
	return ret, err

}

func (v *VoiceSession) StartLoop() {
	var data []byte

	tryReady := 0
	for !v.connection.Ready {
		time.Sleep(1 * time.Second)
		tryReady++
		if tryReady > 30 {
			v.Stop()
			return
		}
	}

	for {
		select {
		case <-v.quit:
			return
		default:
			if len(v.buffer) > 0 {
				v.setSpeaking(true)
				data, v.buffer = v.buffer[0], v.buffer[1:]
				v.connection.OpusSend <- data
			} else {
				v.setSpeaking(false)
				<-v.bufferUpdated
			}
		}
	}
}

func (v *VoiceSession) setSpeaking(speaknig bool) {
	if speaknig != v.speaking {
		if err := v.connection.Speaking(speaknig); err != nil {
			v.Stop()
		}
		v.speaking = speaknig
	}
}

func (v *VoiceSession) SetBuffer(data [][]byte) {
	if v.bot.stopping {
		return
	}
	isZero := len(v.buffer) == 0
	v.buffer = data
	if isZero && len(v.buffer) != 0 {
		v.bufferUpdated <- struct{}{}
	}
}

func (v *VoiceSession) Stop() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error recovered in VoiceSession.Stop() for "+v.connection.GuildID, r)
		}
		if err := v.connection.Disconnect(); err != nil {
			fmt.Println("Unable to disconnect?!", err)
		}
		time.Sleep(50 * time.Millisecond)
		v.connection.Close()
	}()

	// Remove voice session from bot and stop the loop.
	delete(v.bot.Sessions, v.connection.GuildID)
	close(v.quit)

	if v.connection.Ready {
		// Broadcast "Goodbye".
		v.setSpeaking(true)
		for _, data := range goodbye {
			v.connection.OpusSend <- data
		}
		v.setSpeaking(false)
	}
}
