package discordio

import (
	"fmt"
	"io"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type MessageWriter struct {
	io.WriteCloser
	Session   *discordgo.Session
	Message   *discordgo.MessageCreate
	TotalSent uint
	MaxSent   uint
	CodeBlock bool
	Messages  []string
	Size      int
}

func NewMessageWriter(session *discordgo.Session, message *discordgo.MessageCreate) *MessageWriter {
	return &MessageWriter{
		Session:   session,
		Message:   message,
		TotalSent: 0,
		MaxSent:   4,
		CodeBlock: true,
		Messages:  []string{},
		Size:      0,
	}
}

func (w *MessageWriter) Write(p []byte) (n int, err error) {

	// TODO past line no break

	input := string(p[:])
	lines := strings.Split(removeShittyReturns(input), "\n")
	for k, v := range lines {
		if len(v)+w.Size+1 >= 1990 {
			w.sendMessage()
		}

		w.Size += len(v)
		w.Messages = append(w.Messages, v+"\n")
	}

	return len(p), nil

}

func (w *MessageWriter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

func (w *MessageWriter) WriteLine(s string) (n int, err error) {
	return w.Write([]byte(s + "\n"))
}

func (w *MessageWriter) Close() error {
	w.sendMessage()
	return nil
}

func (w *MessageWriter) sendMessage() {

	if w.TotalSent >= w.MaxSent {
		w.Size = 0
		w.Messages = []string{}
		return
	}

	msg := strings.Join(w.Messages, "")
	if msg == "" {
		return
	}

	if w.CodeBlock {
		msg = "```" + strings.Replace(msg, "`", "\\`", -1) + "```"
	}

	_, err := w.Session.ChannelMessageSend(w.Message.ChannelID, msg)
	if err != nil {
		fmt.Println("Error occured: ", err)
	}

	w.Size = 0
	w.Messages = []string{}
	w.TotalSent++

}

func removeShittyReturns(str string) string {
	str = strings.Replace(str, "\r\n", "\n", -1)
	str = strings.Replace(str, "\r", "\n", -1)
	return str
}
