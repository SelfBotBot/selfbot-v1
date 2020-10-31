package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"

	alexa2 "github.com/SelfBotBot/selfbot/alexa"

	"github.com/SelfBotBot/selfbot/data"

	"github.com/SelfBotBot/selfbot/web"

	"github.com/SelfBotBot/selfbot/config"
	"github.com/SelfBotBot/selfbot/discord"
)

func main() {

	address := flag.String("address", "", "The address to listen on (overrides)")
	flag.Parse()

	conf := new(config.Config)
	e(conf.Load())

	panel, err := web.New(conf)
	e(err)

	alexa := new(alexa2.AlexaMeme)
	alexa.Web = panel
	alexa.RegisterHandlers()

	dbEngine, err := data.NewHandler(conf.MySQL)
	e(err)
	panel.Data = dbEngine

	go func() {
		if *address != "" {
			e(panel.Gin.Run(*address))
		} else {
			e(panel.RunAutoTLS())
		}
	}()

	bot, err := discord.New(conf.Discord.Token)
	e(err)

	files, err := ioutil.ReadDir("./")
	e(err)

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".dca") {
			if err := bot.LoadSound(f.Name(), f.Name()[0:len(f.Name())-4]); err != nil {
				fmt.Println("Unable to load "+f.Name()+", ", err)
				continue
			}
		}
	}

	panel.PlaySound = func(guildId, sound string) bool {
		soundData, sok := bot.Sounds[sound]
		fmt.Println("Sound found: ", sound)
		if voiceSession, ok := bot.Sessions[guildId]; sok && ok {
			voiceSession.SetBuffer(soundData)
			return true
		}
		return false
	}

	err = bot.Session.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		panic(err)
	}

	fmt.Println("Done..")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	e(bot.Close())
	os.Exit(0)
}

func e(err error) {
	if err != nil {
		panic(err)
	}
}
