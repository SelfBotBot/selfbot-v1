package discord

import (
	"errors"

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
	close(v.quit)
	delete(v.bot.Sessions, v.connection.GuildID)
	v.connection.Close()
}
