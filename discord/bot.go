package discord

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/SelfBotBot/selfbot/discord/info"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session    *discordgo.Session
	Sounds     map[string][][]byte
	Sessions   map[string]*VoiceSession
	stopping   bool
	infoModule *info.InfoModule
}

func New(token string) (*Bot, error) {
	ret := &Bot{
		Sessions:   make(map[string]*VoiceSession),
		Sounds:     make(map[string][][]byte),
		infoModule: &info.InfoModule{},
	}
	var err error
	ret.Session, err = discordgo.New("Bot " + token)

	ret.Session.AddHandler(ret.botCommand)
	ret.Session.AddHandler(ret.ready)

	return ret, err
}

func (b *Bot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	s.UpdateStatus(0, "/soundboard | /sb")
}

func (b *Bot) FindUserInGuild(UserID string, GuildID string) (ChannelID string, err error) {

	fmt.Println(GuildID)
	g, err := b.Session.State.Guild(GuildID)
	if err != nil {
		fmt.Printf("%#v\n", b.Session.State.Guilds)
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == UserID {
			ChannelID = vs.ChannelID
			return
		}
	}

	err = errors.New("no user in channel")
	return

}

// loadSound attempts to load an encoded sound file from disk.
func (b *Bot) LoadSound(fileName, name string) error {
	data, err := LoadSound(fileName)
	if err != nil {
		return err
	}

	b.Sounds[name] = data
	return nil
}

func LoadSound(fileName string) ([][]byte, error) {

	var ret [][]byte
	file, err := os.Open(fileName)
	if err != nil {
		return ret, err
	}

	var opuslen int16
	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return ret, err
			}
			return ret, nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return ret, err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return ret, err
		}

		ret = append(ret, InBuf)
	}

	return ret, nil

}
func (b *Bot) Close() error {
	b.stopping = true
	for _, v := range b.Sessions {
		v.buffer = goodbye
		v.bufferUpdated <- struct{}{}
	}
	time.Sleep(1 * time.Second)
	for _, v := range b.Sessions {
		v.Stop()
	}
	return b.Session.Close()
}
