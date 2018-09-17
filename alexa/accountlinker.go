package alexa

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/SelfBotBot/selfbot/data"

	"github.com/mikeflynn/go-alexa/skillserver"
)

func (a *AlexaMeme) LinkAccount(echoReq *skillserver.EchoRequest) *skillserver.EchoResponse {

	var latestErr error

	// This is aids, please excuse this...
	colourA, err := getSlotID(echoReq, "COLOUR_A")
	setErrorIfPreviousIsNil(latestErr, err)
	b, err := getSlotID(echoReq, "COLOUR_B")
	setErrorIfPreviousIsNil(latestErr, err)
	c, err := getSlotID(echoReq, "COLOUR_C")
	setErrorIfPreviousIsNil(latestErr, err)
	d, err := getSlotID(echoReq, "COLOUR_D")
	setErrorIfPreviousIsNil(latestErr, err)
	e, err := getSlotID(echoReq, "COLOUR_E")
	setErrorIfPreviousIsNil(latestErr, err)
	if err != nil {
		return skillserver.NewEchoResponse().OutputSpeech("Sorry, we have encountered an error... We can't get all the colours").EndSession(true)
	}

	key := "ALEXA.LINKING:" + colourA + "_" + b + "_" + c + "_" + d + "_" + e
	fmt.Println("KEY is ", key)

	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*6))
	redis, err := a.Web.Redis.GetContext(ctx)
	if err != nil {
		return skillserver.NewEchoResponse().OutputSpeech("Sorry, we have encountered an error... We can't connect to redis.").EndSession(true)
	}

	reply, err := redis.Do("GET", key)
	if err != nil {
		return skillserver.NewEchoResponse().OutputSpeech("We couldn't find your discord account, this has to be done within 5 minutes of generating the linking keys.").EndSession(true)
	}

	redis.Close()
	fmt.Printf("%#v\n", reply)
	discordId, ok := reply.([]byte)
	if !ok {
		return skillserver.NewEchoResponse().OutputSpeech("You need to generate linking keys to use this, go to s b dot cory dot red forward slash alexa link for more information.").EndSession(true)
	}

	userId, err := strconv.ParseUint(string(discordId), 10, 64)
	if err != nil {
		return skillserver.NewEchoResponse().OutputSpeech("We encountered an error processing your discord user ID.").EndSession(true)
	}

	user := &data.User{
		ID: userId,
	}

	// engine create.
	engine := a.Web.Data.Engine
	if err := engine.Where(user).First(user).Error; err != nil || !user.Agreed {
		return skillserver.NewEchoResponse().OutputSpeech("You need to be registered with SelfBot to do this properly. This includes accepting the ToS.").EndSession(true)
	}

	user.AlexaID = echoReq.GetUserID()
	if err := engine.Save(user).Error; err != nil {
		return skillserver.NewEchoResponse().OutputSpeech("Something went wrong updating your user account.").EndSession(true)
	}

	return skillserver.NewEchoResponse().OutputSpeech("Ohh yes honey, this should be done. Please log out and log back in on the website.").EndSession(true)

}

func setErrorIfPreviousIsNil(previousError, newError error) error {
	if previousError == nil {
		return newError
	}

	return previousError
}

func getSlotID(request *skillserver.EchoRequest, name string) (string, error) {
	slot, err := request.GetSlot(name)
	if err != nil {
		return "", err
	}

	id, ok := slot.Resolutions.ResolutionsPerAuthority[0].Values[0]["value"]
	if ok {
		return id.ID, nil
	}

	return slot.Value, nil

}
