package alexa

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/SilverCory/EzVote/pkg/dep/sources/https---github.com-kataras-iris/core/errors"

	"github.com/SelfBotBot/selfbot/web"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	"github.com/gin-gonic/gin"

	alexa "github.com/mikeflynn/go-alexa/skillserver"
)

var SelfbotIdentificationWords = []string{
	"RED", "GREEN", "BLUE", "GOLD", "PINK", "BLACK", "WHITE", "SILVER", "GREY", "BROWN",
}

var SelfbotDoneResponses = []string{
	"Sure.",
	"Lets do it!",
	"Done and Done",
	"All righty.",
	"Fuck yeah.",
	"Woo!",
	"With pleasure.",
	"As you say.",
}

type AlexaMeme struct {
	Web   *web.Panel
	Party *gin.RouterGroup
	web.Handler
}

// Alright what you're about to see here is pretty gross.. Sorry.
func (a *AlexaMeme) RegisterHandlers() error {
	alexa.EchoPrefix = "/alexamemes/echo/"
	router := mux.NewRouter()
	alexa.Init(map[string]interface{}{
		"/alexamemes/echo/selfbot": alexa.EchoApplication{
			AppID:   a.Web.Config.Web.AlexaAppID,
			Handler: a.EchoSelfBot,
		},
	}, router)

	n := negroni.New(negroni.NewRecovery())
	n.UseHandler(router)

	a.Party = a.Web.Gin.Group("/alexamemes/")
	a.Party.Any("echo/*dab1", func(ctx *gin.Context) {
		n.ServeHTTP(ctx.Writer, ctx.Request)
	})

	a.Web.Gin.GET("/alexalink", a.linkAccount)

	return nil
}

func (a *AlexaMeme) linkAccount(ctx *gin.Context) {
	user, ok := a.Web.GetUser(ctx)
	if !ok || user.CreatedAt.IsZero() {
		ctx.AbortWithError(401, errors.New("unauthorised? please log in"))
		return
	}

	redis, err := a.Web.Redis.GetContext(ctx)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	identWords := ""
	for i := 0; i < 5; i++ {
		identWords += getRandom(SelfbotIdentificationWords) + " "
	}

	key := strings.Replace("ALEXALINKING."+identWords, " ", "_", -1)
	if err := redis.Send("SETEX", key, 600, user.ID); err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	ctx.String(200, identWords)

}

func (a *AlexaMeme) EchoSelfBot(w http.ResponseWriter, r *http.Request) {
	echoReq := alexa.GetEchoRequest(r)

	if echoReq.GetRequestType() == "IntentRequest" {
		var echoResp *alexa.EchoResponse
		switch echoReq.GetIntentName() {
		case "play":

			var soundName string

			slot, ok := echoReq.AllSlots()["sound"]
			if len(slot.Resolutions.ResolutionsPerAuthority) > 0 && slot.Resolutions.ResolutionsPerAuthority[0].Status.Code == "ER_SUCCESS_MATCH" {
				id, ok := slot.Resolutions.ResolutionsPerAuthority[0].Values[0]["value"]
				if ok {
					soundName = id.ID
				}
			} else if ok {
				soundName = a.transformName(slot.Value)
			}

			if a.Web.PlaySound != nil && a.Web.PlaySound("217977786248331274", soundName) {
				echoResp = alexa.NewEchoResponse().OutputSpeech(getRandom(SelfbotDoneResponses)).EndSession(false)
			} else {
				echoResp = alexa.NewEchoResponse().OutputSpeech("I'm sorry, I couldn't find the sound " + soundName + ", try again?").EndSession(false)
			}
			break
		case "LinkAccount":
			echoResp = a.LinkAccount(echoReq)
		case "CancelIntent":
			echoResp = alexa.NewEchoResponse().OutputSpeech("Cya").EndSession(true)
			break

		case "StopIntent":
			a.Web.PlaySound("402871667891765248", "oof")
			echoResp = alexa.NewEchoResponse().OutputSpeech("I hope I stopped in time!").EndSession(true)
			break

		default:
			echoResp = alexa.NewEchoResponse().OutputSpeech("I'm sorry, I didn't get that. Can you say that again?").EndSession(false)
		}

		json, _ := echoResp.String()
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write(json)
	} else if echoReq.GetRequestType() == "SessionEndedRequest" {
		//session.Delete(col)
	}
}

func (a *AlexaMeme) transformName(value string) string {
	return strings.Replace(strings.ToLower(value), " ", "_", -1)
}

func getRandom(list []string) string {
	return list[rand.Intn(len(list))]
}
