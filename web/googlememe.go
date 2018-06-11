package web

import (
	"github.com/crezam/actions-on-google-golang/model"
	"github.com/gin-gonic/gin"
)

type GoogleMeme struct {
	Web   *Panel
	Party *gin.RouterGroup
	Handler
}

// Alright what you're about to see here is pretty gross.. Sorry.
func (g *GoogleMeme) RegisterHandlers() error {
	g.Party = g.Web.Gin.Group("/googlememes/")
	g.Party.POST("/webhook", g.HandleWebhook)

	return nil
}

func (g *GoogleMeme) HandleWebhook(ctx *gin.Context) {
	var webhookRequest model.ApiAiRequest
	ctx.BindJSON(&webhookRequest)
	// TODO
}
