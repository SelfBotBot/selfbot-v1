package discord

import (
	"sort"
	"strings"

	"github.com/SelfBotBot/selfbot/discord/discordio"
	"github.com/bwmarrin/discordgo"
)

func (b *Bot) botCommandListener(s *discordgo.Session, m *discordgo.MessageCreate) {

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

	if strings.EqualFold(m.Content, "/leave") {
		b.leaveCommand(s, m, c, g)
	} else if strings.EqualFold(m.Content, "/sounds") {
		b.soundsCommand(s, m, c, g)
	} else if strings.EqualFold(m.Content, "/join") {
		b.joinCommand(s, m, c, g)
	} else if strings.HasPrefix(m.Content, "/play ") {
		b.playCommand(s, m, c, g)
	} else if strings.HasPrefix(m.Content, "/info") {
		b.infoCommand(s, m, c, g)
	} else if strings.HasPrefix(m.Content, "/sb") || strings.HasPrefix(m.Content, "/soundboard") {
		s.ChannelMessageSend(c.ID, "https://sb.cory.red/panel/"+c.GuildID)
		s.ChannelMessageSend(c.ID, "We're still being made, but you can use `/join` to make me join your voice channel,\nand `/play [sound]` to play the audio file once it's in there!\nYou can also list the available sounds using `/sounds`.")
		return
	}

}

func (b *Bot) joinCommand(s *discordgo.Session, m *discordgo.MessageCreate, c *discordgo.Channel, g *discordgo.Guild) {
	channel, err := b.FindUserInGuild(m.Author.ID, g.ID)
	if err != nil {
		s.ChannelMessageSend(c.ID, "Unable to find you in VC.\n```"+err.Error()+"```")
	}

	vs, err := NewVoice(s, b, g.ID, channel)
	if err != nil {
		s.ChannelMessageSend(c.ID, "Unable to join VC.\n```"+err.Error()+"```")
		vs.Stop() // Call this incase we're still alive?
		return
	}

	go vs.StartLoop()
	b.Sessions[g.ID] = vs
	go vs.SetBuffer(welcome)
	return
}

func (b *Bot) leaveCommand(s *discordgo.Session, m *discordgo.MessageCreate, c *discordgo.Channel, g *discordgo.Guild) {
	if k, v := b.Sessions[g.ID]; v {
		k.Stop()
	}
}

func (b *Bot) playCommand(s *discordgo.Session, m *discordgo.MessageCreate, c *discordgo.Channel, g *discordgo.Guild) {
	sound, ok := b.Sounds[m.Content[6:]]
	if !ok {
		s.ChannelMessageSend(c.ID, "No such sound exists! Usage `/play [sound]`")
		return
	}

	ses, ok := b.Sessions[g.ID]
	if ok {
		ses.SetBuffer(sound)
	}
	go s.ChannelMessageDelete(c.ID, m.ID) // Delete /play message to reduce spam
	return
}

func (b *Bot) soundsCommand(s *discordgo.Session, m *discordgo.MessageCreate, c *discordgo.Channel, g *discordgo.Guild) {

	// TODO make this way better, sorting this every time is ew.
	keys := make([]string, 0, len(b.Sounds))
	for k := range b.Sounds {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	writer := discordio.NewMessageWriter(s, m)
	writer.CodeBlock = false

	writer.Write([]byte("Here's a list of available sounds!\n"))
	for _, v := range keys {
		writer.Write([]byte("`/play " + v + "`\n"))
	}

	writer.Close()
}

func (b *Bot) infoCommand(s *discordgo.Session, m *discordgo.MessageCreate, c *discordgo.Channel, g *discordgo.Guild) {
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
