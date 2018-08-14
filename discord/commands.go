package discord

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

		vs, err := NewVoice(s, b, g.ID, channel)
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
			ses.SetBuffer(sound)
		}
		return
	}

	if strings.HasPrefix(m.Content, "/info") {
		if len(m.Content) <= 5 {
			b.infoModule.AllStatsCommand(s, m)
			return
		} else {
			startsWith := strings.TrimSpace(m.Content[6:])
			if strings.HasPrefix(startsWith, "h") {
				b.infoModule.HostStatsCommand(s, m)
			} else if strings.HasPrefix(startsWith, "b") {
				b.infoModule.BotStatsCommand(s, m)
			} else if strings.HasPrefix(startsWith, "a") {
				b.infoModule.AllStatsCommand(s, m)
			} else {
				s.ChannelMessageSend(c.ID, "Hey, you need to /stats [all|bot|host]")
			}
			return
		}
	}

	if strings.HasPrefix(m.Content, "/sb") || strings.HasPrefix(m.Content, "/soundboard") {
		s.ChannelMessageSend(c.ID, "https://sb.cory.red/panel/"+c.GuildID)
		s.ChannelMessageSend(c.ID, "We're still being made, but you can use `/join` to make me join your voice channel,\nand `/play [sound]` to play the audio file once it's in there!\nYou can also list the available sounds using `/sounds`.")
		return
	}

}
