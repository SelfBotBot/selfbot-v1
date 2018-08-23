package web

import (
	"github.com/SelfBotBot/selfbot/web/viewdata"
	"github.com/gin-gonic/gin"
)

const PrivacyPolicy = `
<html><head><title>Privacy</title></head>
<body>
	<h1>privacy policy</h1>
	<p>We aren't currently logging any information as of yet however in the near future, your google account, amazon account and discord account informations will be stored and used for the usage of this.</p>
</body>
</html>'`

type Pages struct {
	Web *Panel
	Handler
}

func (p *Pages) RegisterHandlers() error {

	p.Web.Gin.GET("/", p.handleIndex)
	p.Web.Gin.GET("/tos", p.handleTos)

	return nil
}

func (p *Pages) handlePanel(ctx *gin.Context) {
	//sess := sessions.Default(ctx)
	v := viewdata.Default(ctx)
	v.Set("Title", "Panel")
	v.HTML(200, "pages/index.html")
}

func (p *Pages) handleIndex(ctx *gin.Context) {
	v := viewdata.Default(ctx)
	v.Set("Title", "Home")
	v.HTML(200, "pages/index.html")
}

func (p *Pages) handleTos(ctx *gin.Context) {
	v := viewdata.Default(ctx)
	v.Set("Title", "Terms Of Service")
	v.HTML(200, "pages/tos.html")
}

func (p *Pages) handlePrivacyPolicy(ctx *gin.Context) {
	ctx.String(200, PrivacyPolicy)
}
