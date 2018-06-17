package web

import (
	"errors"
	"strconv"

	"github.com/SelfBotBot/selfbot/data"

	"fmt"

	"os"

	"github.com/SelfBotBot/selfbot/web/viewdata"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
)

type Oauth struct {
	Web   *Panel
	Party *gin.RouterGroup
	Handler
}

func (o *Oauth) RegisterHandlers() error {

	discordConf := o.Web.Config.DiscordOAuth
	goth.UseProviders(
		discord.New(discordConf.Key, discordConf.Secret, discordConf.Callback+"/auth/discord/callback", discord.ScopeEmail, discord.ScopeIdentify, discord.ScopeGuilds, discord.ScopeJoinGuild),
	)

	o.Web.Gin.GET("/login", o.handleLogin)
	o.Web.Gin.GET("/logout", o.handleLogout)
	o.Web.Gin.GET("/register", o.handleRegisterGet)
	o.Web.Gin.POST("/register", o.handleRegisterPost)

	o.Party = o.Web.Gin.Group("/auth/discord/")
	o.Party.GET("/", o.handleIndex)
	o.Party.GET("/callback", o.handleCallback)

	return nil
}

func (o *Oauth) handleLogout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Clear()
	if err := SaveSession(sess, ctx); err != nil {
		ctx.Error(err)
		return
	}
	ctx.Redirect(302, "/")
	ctx.Next()
}

func (o *Oauth) handleIndex(ctx *gin.Context) {
	ctx.String(200, "<h1>Hello from the auth/discord \\o/</h1>\n"+
		"<p>I put a typo there so you KNOW that it's not just spitting out the path..</p>")
	ctx.Next()
}

func (o *Oauth) handleLogin(ctx *gin.Context) {

	redirectTo := ctx.Param("redirectTo")
	if redirectTo == "" {
		redirectTo = "/"
	}

	sess := sessions.Default(ctx)
	if sess.Get(SessionAuthKey) != nil {
		ctx.Redirect(302, redirectTo)
		ctx.Next()
		return
	}

	provider, _ := goth.GetProvider("discord")
	oauthSess, err := provider.BeginAuth(ctx.Param("state"))
	if err != nil {
		ctx.Abort()
		return
		// TODO proper error handling.
	}

	url, err := oauthSess.GetAuthURL()
	if err != nil {
		ctx.Abort()
		return
		// TODO proper error handling.
	}

	sess.Set(SessionAuthKey, oauthSess.Marshal())
	sess.Set(SessionRedirectKey, redirectTo)
	if err := SaveSession(sess, ctx); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Redirect(302, url)
	ctx.Next()

}

func (o *Oauth) handleCallback(ctx *gin.Context) {

	provider, err := goth.GetProvider("discord")
	if err != nil {
		ctx.Abort()
		return
		// TODO proper error handling.
	}

	sess := sessions.Default(ctx)
	if sess.Get(SessionAuthKey) == nil {
		ctx.Error(errors.New("completeUserAuth error: could not find a matching session for this request"))
		return
		// TODO proper error handling.
	}

	oauthSess, err := provider.UnmarshalSession(sess.Get(SessionAuthKey).(string))
	if err != nil {
		ctx.Error(fmt.Errorf("completeUserAuth error: could not unmarshal session data. Error: %#v", err))
		return
		// TODO proper error handling.
	}

	_, err = oauthSess.Authorize(provider, ctx.Request.URL.Query())
	if err != nil {
		ctx.Error(fmt.Errorf("completeUserAuth error: bad auth! Error: %#v", err))
		return
		// TODO proper error handling.
	}

	u, err := provider.FetchUser(oauthSess)

	userId, err := strconv.ParseUint(u.UserID, 10, 64)
	if err != nil {
		ctx.Abort()
		ctx.Error(fmt.Errorf("completeuserAuth! Error: %#v", err))
	}

	// Query data
	user := &data.User{
		ID:    userId,
		Email: u.Email,
	}

	if _, bol := os.LookupEnv("OAUTHTEST"); bol {
		fmt.Printf("%#v\n", u)
	}

	// engine create.
	engine := o.Web.Data.Engine
	engine.Where(user).First(user)

	user.Name = u.Name
	user.Discriminator = u.RawData["discriminator"].(string)
	user.SessionToken = u.AccessToken
	user.SessionTokenSecret = u.AccessTokenSecret
	user.RefreshToken = u.AccessToken
	user.Expiry = u.ExpiresAt

	if user.CreatedAt.IsZero() {
		fmt.Println("Create")
		engine.Create(user)
	} else {
		fmt.Println("Save")
		engine.Save(user)
	}

	// User authenticated.
	sess.Set("user", *user)
	if err := SaveSession(sess, ctx); err != nil {
		ctx.Error(err)
		return
	}
	fmt.Printf("User authenticate via discord! %q %s#%s", u.UserID, u.Name, u.RawData["discriminator"])

	// Redirect to the appropriate place.
	redirectTo := sess.Get(SessionRedirectKey).(string)
	if redirectTo == "" {
		redirectTo = "/"
	}

	if !user.Agreed {
		redirectTo = "/register"
	}

	ctx.Redirect(302, redirectTo)
	ctx.Next()

}

func (o *Oauth) handleRegisterGet(ctx *gin.Context) {
	if o.GetUserOrRedirect(ctx).Expiry.IsZero() {
		return
	}
	v := viewdata.Default(ctx)
	v.Set("Title", "Registration")
	v.HTML(200, "pages/register.html")
}

func (o *Oauth) handleRegisterPost(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	user := o.GetUserOrRedirect(ctx)
	if user.Expiry.IsZero() {
		return
	}

	redirectTo := ""
	if redirectTo = sess.Get(SessionRedirectKey).(string); redirectTo == "" {
		redirectTo = "/"
	}
	sess.Delete(SessionRedirectKey)

	accept := ctx.PostForm("accept")
	if accept != "accept" {
		sess.Clear()
		if err := SaveSession(sess, ctx); err != nil {
			ctx.Error(err)
			return
		}
		ctx.Redirect(302, "/")
		ctx.Next()
		return
	}

	// Update the user.
	user.Agreed = true
	o.Web.Data.Engine.Save(&user)
	sess.Set("user", user)
	if err := SaveSession(sess, ctx); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Redirect(302, redirectTo)
	ctx.Next()

}

func (o *Oauth) GetUserOrRedirect(ctx *gin.Context) data.User {
	sess := sessions.Default(ctx)
	user, ok := sess.Get("user").(data.User)
	if !ok || user.Agreed {
		if redirectTo := sess.Get(SessionRedirectKey); redirectTo != nil { // TODO remove key
			if to, ok := redirectTo.(string); ok {
				ctx.Redirect(302, to)
				ctx.Next()
				return data.User{}
			}
		}
		ctx.Redirect(302, "/")
		ctx.Next()
		return data.User{}
	}

	return user
}
