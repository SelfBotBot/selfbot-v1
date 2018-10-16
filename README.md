# selfbot
A Golang Discord soundboard bot.

## Information
The ultimate soundboard bot with plenty of planned features.

SelfBot is simple to use, easy to add custom sounds to, with a simple and ergonomic web interface.

## TODO
- [ ] Finish web UI
- [ ] Add proper support for sharding
- [ ] Redo audio handler (maybe); use locks and what not instead of channels.
- [ ] Write apps and programs for hotkey/mobile phone support.
- [x] Alexa integration
  - [ ] Publish alexa integration
- [ ] Google assistant integration

## Self hosting
Not sure how well this will work, some code might need changing due to SSL and stuff..

### Requirments
- Golang
- dep
- FFmpeg
- Redis
- MariaDB/MySQL

### Execution/build
The easiest way to run this program is to use the golang path and git, that way you won't have to worry about all of the assets and audio file locations.
1. `mkdir "$GOPATH/SelfBotBot/selfbot"`
2. `git clone https://github.com/SelfBotBot/selfbot.git "$GOPATH/SelfBotBot/selfbot"`
3. `cd "$GOPATH/SelfBotBot/selfbot"`
4. `dep ensure`
5. `go run cmd/selfbot/main.go`

A configuration file will be created and the program will exit and ask for it to be edited. Fill in the apropriate information and restart the bot.

