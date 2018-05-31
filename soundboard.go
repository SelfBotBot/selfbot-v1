package main

import (
	"bufio"
	"fmt"
	"github.com/SilverCory/soundboard/discord"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func main() {

	config := &Config{}
	e(config.Load())

	bot, err := discord.New(config.Discord.Token)
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

	err = bot.Session.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		panic(err)
	}

	fmt.Println("Done..")

	reader := bufio.NewReader(os.Stdin)
	var guild string
	sound := "distortion"
	for {

		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if strings.HasPrefix(string(line), "join ") {
			args := strings.Split(string(line[5:]), " ")
			if len(args) != 2 {
				fmt.Println("Bad args. Need \"join [guild] [channel]\"")
				continue
			}
			guild = args[0]

			go func() {
				voiceSesh, err := discord.NewVoice(bot.Session, args[0], args[1])
				if err != nil {
					fmt.Println("Unable to join voice??", err)
					return
				}

				bot.Sessions[args[0]] = voiceSesh
				time.Sleep(250 * time.Millisecond)
				go voiceSesh.StartLoop()
			}()

			continue

		} else if strings.HasPrefix(string(line), "find ") {
			args := strings.Split(string(line[5:]), " ")
			if len(args) != 2 {
				fmt.Println("Bad args. Need \"find [guild] [userid]\"")
				continue
			}
			fmt.Println(bot.FindUserInGuild(args[1], args[0]))
			continue

		} else if strings.HasPrefix(string(line), "choose ") {
			args := strings.Split(string(line[7:]), " ")
			if len(args) != 1 {
				fmt.Println("Bad args. Need \"choose [guild]\"")
				continue
			}
			guild = args[0]
			continue

		} else if strings.HasPrefix(string(line), "sound ") {
			args := line[6:]
			if len(args) == 0 {
				fmt.Println("Bad args. Need \"sound [sound]\"")
				continue
			}
			sound = string(args)
			continue

		} else {
			ses, ok := bot.Sessions[guild]
			if ok {
				ses.Buffer = bot.Sounds[sound]
			}
		}

	}

}

func e(err error) {
	if err != nil {
		panic(err)
	}
}
