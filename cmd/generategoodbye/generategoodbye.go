package main

import (
	"fmt"

	"github.com/SelfBotBot/selfbot/discord"
)

func main() {
	data, err := discord.LoadSound("goodbye.dca")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", data)

}
