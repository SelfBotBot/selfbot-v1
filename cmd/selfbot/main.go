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

	conf := &config.Config{}
	e(conf.Load())

	panel, err := web.New(conf)
	e(err)

	alexa := &alexa2.AlexaMeme{Web: panel}
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

	//reader := bufio.NewReader(os.Stdin)
	//var guild string
	//sound := "distortion"
	//for {
	//
	//	line, _, err := reader.ReadLine()
	//	if err != nil {
	//		fmt.Println(err)
	//		continue
	//	}
	//
	//	if strings.HasPrefix(string(line), "join ") {
	//		args := strings.Split(string(line[5:]), " ")
	//		if len(args) != 2 {
	//			fmt.Println("Bad args. Need \"join [guild] [channel]\"")
	//			continue
	//		}
	//		guild = args[0]
	//
	//		go func() {
	//			voiceSesh, err := discord.NewVoice(bot.Session, args[0], args[1])
	//			if err != nil {
	//				fmt.Println("Unable to join voice??", err)
	//				return
	//			}
	//
	//			bot.Sessions[args[0]] = voiceSesh
	//			time.Sleep(250 * time.Millisecond)
	//			go voiceSesh.StartLoop()
	//		}()
	//
	//		continue
	//
	//	} else if strings.HasPrefix(string(line), "find ") {
	//		args := strings.Split(string(line[5:]), " ")
	//		if len(args) != 2 {
	//			fmt.Println("Bad args. Need \"find [guild] [userid]\"")
	//			continue
	//		}
	//		fmt.Println(bot.FindUserInGuild(args[1], args[0]))
	//		continue
	//
	//	} else if strings.HasPrefix(string(line), "choose ") {
	//		args := strings.Split(string(line[7:]), " ")
	//		if len(args) != 1 {
	//			fmt.Println("Bad args. Need \"choose [guild]\"")
	//			continue
	//		}
	//		guild = args[0]
	//		continue
	//
	//	} else if strings.HasPrefix(string(line), "sound ") {
	//		args := line[6:]
	//		if len(args) == 0 {
	//			fmt.Println("Bad args. Need \"sound [sound]\"")
	//			continue
	//		}
	//		sound = string(args)
	//		continue
	//
	//	} else if strings.HasPrefix(string(line), "load ") {
	//		args := line[5:]
	//		if len(args) == 0 {
	//			fmt.Println("Bad args. Need \"load [file]\"")
	//			continue
	//		}
	//		sound = string(args)
	//		if err := bot.LoadSound(string(args), string(args[0:len(args)-4])); err != nil {
	//			fmt.Println("Error loading file! ", err)
	//			continue
	//		}
	//		continue
	//
	//	} else {
	//		ses, ok := bot.Sessions[guild]
	//		if ok {
	//			ses.SetBuffer(bot.Sounds[sound])
	//		}
	//	}
	//
	//}

}

func e(err error) {
	if err != nil {
		panic(err)
	}
}
