package discord

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"

	_ "github.com/shirou/w32"
)

const (
	BlankText      = "\u200B"
	HorizontalRule = "──────────────────────────────────────────"
)
const (
	ColourGood = 0x73C62F
	ColourEh   = 0xE2D044
	ColourBad  = 0xE05545
)

const (
	RuntimeInfoTitles = "Version: \nGOARCH: "
	RuntimeInfo       = "%s\n%s"

	CPUInfoTitles = "No GoRoutines: \nNo CGo Calls: "
	CPUInfo       = "%d\n%d"

	GCInfoTitles = "GC Runs: \nGC System: "
	GCInfo       = "%d\n%s"

	HeapInfoTitles = "Heap Usage: \nHeap In Use: \nHeap Objects: "
	HeapInfo       = "%s/%s\n%s\n%d"

	SysInfoTitles = "Total Used: \nTotal GC'd: "
	SysInfo       = "%s\n%s"

	ProcInfoErrTitles = "Total Procs: "
	ProcInfoErr       = "%d"

	ProcInfoTitles = "Total Procs: \nModel Name: \nSpeed"
	ProcInfo       = "%d\n%s\n%.2fMhz"

	HostInfoTitles = "OS: \nPlatform: \nVirtualization: \nRole: \nUptime: "
	HostInfo       = "%s\n%s\n%s\n%s\n%s"
)

type InfoModule struct{}

func (p *InfoModule) allStatsCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	p.React(s, m, "⏳")
	embedBad := p.buildEmbed(p.populateMemStats(), true)
	embedBad.Author = &discordgo.MessageEmbedAuthor{
		Name: "Bot Stats.",
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embedBad)
	if err != nil {
		fmt.Println(err)
	}

	p.Unreact(s, m, "⏳")
	p.React(s, m, "✅")
}

func (p *InfoModule) botStatsCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	p.React(s, m, "⏳")
	embedBad := p.buildEmbed(p.populateMemStats(), false)
	embedBad.Author = &discordgo.MessageEmbedAuthor{
		Name: "Bot Runtime Stats.",
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embedBad)
	if err != nil {
		fmt.Println(err)
	}

	p.Unreact(s, m, "⏳")
	p.React(s, m, "✅")
}

func (p *InfoModule) hostStatsCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	p.React(s, m, "⏳")
	embedBad := p.buildEmbed(nil, true)
	embedBad.Author = &discordgo.MessageEmbedAuthor{
		Name: "Bot Host Stats.",
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embedBad)
	if err != nil {
		fmt.Println(err)
	}

	p.Unreact(s, m, "⏳")
	p.React(s, m, "✅")
}

func (p *InfoModule) populateMemStats() *runtime.MemStats {
	stats := &runtime.MemStats{}
	runtime.ReadMemStats(stats)
	return stats
}

func (p *InfoModule) buildEmbed(memStats *runtime.MemStats, system bool) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField
	colour := ColourGood

	// Runtime section
	if memStats != nil {

		// Horizontal rule for separation.
		fields = append(fields, RuleField())

		// Runtime
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Runtime Information",
			Inline: true,
			Value:  RuntimeInfoTitles,
		})
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   BlankText,
			Inline: true,
			Value:  fmt.Sprintf(RuntimeInfo, runtime.Version(), runtime.GOARCH),
		})

		// Separator field
		fields = append(fields, SeperatorField())

		// CPU
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "CPU Usage Information",
			Inline: true,
			Value:  CPUInfoTitles,
		})
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   BlankText,
			Inline: true,
			Value:  fmt.Sprintf(CPUInfo, runtime.NumGoroutine(), runtime.NumCgoCall()),
		})

		// Separator field
		fields = append(fields, SeperatorField())

		// GC
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "GC Information",
			Inline: true,
			Value:  GCInfoTitles,
		})
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   BlankText,
			Inline: true,
			Value:  fmt.Sprintf(GCInfo, memStats.NumGC, ByteString(memStats.GCSys)),
		})

		// Separator field
		fields = append(fields, SeperatorField())

		// Heap
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Heap Information",
			Inline: true,
			Value:  HeapInfoTitles,
		})
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   BlankText,
			Inline: true,
			Value:  fmt.Sprintf(HeapInfo, ByteString(memStats.HeapAlloc), ByteString(memStats.HeapSys), ByteString(memStats.HeapInuse), memStats.HeapObjects),
		})

		// Separator field
		fields = append(fields, SeperatorField())

		// System
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Memory Information",
			Inline: true,
			Value:  SysInfoTitles,
		})
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   BlankText,
			Inline: true,
			Value:  fmt.Sprintf(SysInfo, ByteString(memStats.Sys), ByteString(memStats.TotalAlloc)),
		})

		// Separator field
		fields = append(fields, SeperatorField())

		if runtime.NumGoroutine() > 20 || memStats.HeapAlloc >= uint64(float64(memStats.HeapSys)*0.70) {
			colour = ColourEh
			fmt.Println("EH!")
			if memStats.HeapAlloc >= uint64(float64(memStats.HeapSys)*0.80) {
				colour = ColourBad
				fmt.Println("BAAD")
			}
		}

	}

	// System section!
	if system {

		// Horizontal rule for separation.
		fields = append(fields, RuleField())

		cpuInfo, err := cpu.Info()
		if err != nil {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Processor Info",
				Inline: true,
				Value:  ProcInfoErrTitles,
			})
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   BlankText,
				Inline: true,
				Value:  fmt.Sprintf(ProcInfoErr, runtime.NumCPU()),
			})
		} else {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Processor Info",
				Inline: true,
				Value:  ProcInfoTitles,
			})
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   BlankText,
				Inline: true,
				Value:  fmt.Sprintf(ProcInfo, runtime.NumCPU(), cpuInfo[0].ModelName, cpuInfo[0].Mhz),
			})
		}

		// Separator field
		fields = append(fields, SeperatorField())

		hostInfo, err := host.Info()
		if err == nil {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Host Info",
				Inline: true,
				Value:  HostInfoTitles,
			})
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   BlankText,
				Inline: true,
				Value:  fmt.Sprintf(HostInfo, hostInfo.OS, hostInfo.Platform, hostInfo.VirtualizationSystem, hostInfo.VirtualizationRole, time.Duration(int64(hostInfo.Uptime))*time.Second),
			})

			// Separator field
			fields = append(fields, SeperatorField())
		}

	}

	return &discordgo.MessageEmbed{
		Fields: fields,
		Type:   "rich",
		Color:  colour,
	}

}

func RuleField() *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Inline: false,
		Value:  BlankText,
		Name:   HorizontalRule,
	}
}

func SeperatorField() *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Inline: true,
		Value:  BlankText,
		Name:   BlankText,
	}
}

func ByteString(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	floatBytes := float64(bytes)
	exp := uint64(math.Log(floatBytes) / math.Log(1024))
	pre := "KMGTPE"[exp-1]
	return fmt.Sprintf("%.2f %ciB", floatBytes/math.Pow(1024, float64(exp)), pre)

}

func (p *InfoModule) React(s *discordgo.Session, m *discordgo.MessageCreate, emoji string) {
	err := s.MessageReactionAdd(m.ChannelID, m.ID, emoji)
	if err != nil {
		fmt.Println("Error reacting emoji!", err)
	}
}

func (p *InfoModule) Unreact(s *discordgo.Session, m *discordgo.MessageCreate, emoji string) {
	err := s.MessageReactionRemove(m.ChannelID, m.ID, emoji, s.State.User.ID)
	if err != nil {
		fmt.Println("Error unreacting emoji!", err)
	}
}
