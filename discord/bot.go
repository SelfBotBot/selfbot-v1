package discord

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"os"
	"strings"
	"time"
)

type Bot struct {
	Session  *discordgo.Session
	Sounds   map[string][][]byte
	Sessions map[string]*VoiceSession
}

func New(token string) (*Bot, error) {
	ret := &Bot{
		Sessions: make(map[string]*VoiceSession),
		Sounds:   make(map[string][][]byte),
	}
	var err error
	ret.Session, err = discordgo.New("Bot " + token)

	//ret.Session.LogLevel = 3

	ret.Session.AddHandler(ret.botCommand)
	ret.Session.AddHandler(ret.create)
	ret.Session.AddHandler(ret.ready)

	return ret, err
}

func (b *Bot) create(s *discordgo.Session, event *discordgo.GuildCreate) {
	fmt.Println("GuildCreate for " + event.ID)
}

func (b *Bot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	s.UpdateStatus(0, "/soundboard | /sb")
}

func (b *Bot) botCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return
	}

	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		return
	}

	//if strings.EqualFold(m.Content, "/command") {
	//	channel, err := b.FindUserInGuild(m.Author.ID, g.ID)
	//	s.ChannelMessageSend(c.ID, "`find " + g.ID + " " + m.Author.ID +"`")
	//	if err != nil {
	//		s.ChannelMessageSend(c.ID, "Unable to find you in VC.\n```" + err.Error() + "```")
	//	} else {
	//		s.ChannelMessageSend(c.ID, "`join " + g.ID + " " + channel + "`")
	//	}
	//	return
	//}
	if strings.EqualFold(m.Content, "/sounds") {
		msg := "Here's a list of available sounds!\n"
		for k := range b.Sounds {
			msg += "`/play " + k + "`\n"
		}
		s.ChannelMessageSend(c.ID, msg)
		return
	}

	if strings.EqualFold(m.Content, "/join") {
		channel, err := b.FindUserInGuild(m.Author.ID, g.ID)
		if err != nil {
			s.ChannelMessageSend(c.ID, "Unable to find you in VC.\n```"+err.Error()+"```")
		}

		vs, err := NewVoice(s, g.ID, channel)
		if err != nil {
			s.ChannelMessageSend(c.ID, "Unable to join VC.\n```"+err.Error()+"```")
			return
		}

		b.Sessions[g.ID] = vs
		time.Sleep(250 * time.Millisecond)

		go vs.StartLoop()
		return

	}

	if strings.HasPrefix(m.Content, "/play ") {
		sound, ok := b.Sounds[m.Content[6:]]
		if !ok {
			s.ChannelMessageSend(c.ID, "No such sound exists! Usage `/play [sound]`")
			return
		}

		ses, ok := b.Sessions[g.ID]
		if ok {
			ses.Buffer = sound
		}
		return
	}

	if strings.HasPrefix(m.Content, "/sb") || strings.HasPrefix(m.Content, "/soundboard") {
		s.ChannelMessageSend(c.ID, "https://sb.cory.red/panel/"+c.GuildID)
		s.ChannelMessageSend(c.ID, "We're still being made, but you can use `/join` to make me join your voice channel,\nand `/play [sound]` to play the audio file once it's in there!\nYou can also list the available sounds using `/sounds`.")
		return
	}

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

type VoiceSession struct {
	connection *discordgo.VoiceConnection
	Buffer     [][]byte
	quit       chan struct{}
	speaking   bool
}

func NewVoice(s *discordgo.Session, GuildID string, ChannelID string) (*VoiceSession, error) {
	var err error
	ret := &VoiceSession{
		Buffer: make([][]byte, 0),
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
			if len(v.Buffer) > 0 {
				v.setSpeaking(true)
				data, v.Buffer = v.Buffer[0], v.Buffer[1:]
				v.connection.OpusSend <- data
			} else {
				v.setSpeaking(false)
			}
		}
	}
}

func (v *VoiceSession) setSpeaking(speaknig bool) {
	if speaknig != v.speaking {
		v.connection.Speaking(speaknig)
		v.speaking = speaknig
	}
}

func (v *VoiceSession) Stop() {
	close(v.quit)
	v.connection.Close()
}

// loadSound attempts to load an encoded sound file from disk.
func (b *Bot) LoadSound(fileName, name string) error {

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		b.Sounds[name] = append(b.Sounds[name], InBuf)
	}
}
