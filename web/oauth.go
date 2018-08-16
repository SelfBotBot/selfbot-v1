package web

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/SelfBotBot/selfbot/data"

	"fmt"

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

var discordProvider goth.Provider

func (o *Oauth) RegisterHandlers() error {

	discordConf := o.Web.Config.DiscordOAuth
	goth.UseProviders(
		discord.New(discordConf.Key, discordConf.Secret, discordConf.Callback+"/auth/discord/callback", discord.ScopeEmail, discord.ScopeIdentify, discord.ScopeGuilds, discord.ScopeJoinGuild),
	)

	if prov, err := goth.GetProvider("discord"); err != nil {
		panic(err)
	} else {
		discordProvider = prov
	}

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
	if redirectTo == "" || !strings.HasPrefix(redirectTo, "/") {
		redirectTo = "/"
	}

	sess := sessions.Default(ctx)
	if sess.Get(SessionAuthKey) != nil && sess.Get("user") != nil {

		u, err := getAuthUserFromSession(ctx)
		if err == nil && o.refreshToken(u) == nil {
			user, ok := o.GetUser(ctx)
			if !ok {
				ctx.Error(errors.New("man something bad happened here"))
				return
			}

			user.Name = u.Name
			user.Discriminator = u.RawData["discriminator"].(string)
			user.SessionToken = u.AccessToken
			user.SessionTokenSecret = u.AccessTokenSecret
			user.RefreshToken = u.RefreshToken
			user.Expiry = u.ExpiresAt

			if err := o.SaveUser(ctx, &user); err != nil {
				ctx.Error(err)
				return
			}
		}

		ctx.Redirect(302, redirectTo)
		ctx.Next()
		return
	}

	oauthSess, err := discordProvider.BeginAuth(ctx.Param("state"))
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

	sess := sessions.Default(ctx)
	oauthSess, err := getAuthSessionFromSession(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("completeUserAuth error: invalid session! Error: %#v", err))
		return
		// TODO proper error handling.
	}

	_, err = oauthSess.Authorize(discordProvider, ctx.Request.URL.Query())
	if err != nil {
		ctx.Error(fmt.Errorf("completeUserAuth error: bad auth! Error: %#v", err))
		return
		// TODO proper error handling.
	}

	u, err := discordProvider.FetchUser(oauthSess)
	if err != nil {
		ctx.Error(fmt.Errorf("completeUserAuth error: bad user! Error: %#v", err))
		return
		// TODO proper error handling.
	}

	user, err := o.CreateOrGetUser(ctx, u)
	if err != nil {
		ctx.Error(fmt.Errorf("completeUserAuth error: db write/read went bad! Error: %#v", err))
		return
	}

	// Redirect to the appropriate place.
	redirectTo := sess.Get(SessionRedirectKey).(string)
	if !user.Agreed {
		redirectTo = "/register"
	} else {
		if redirectTo == "" {
			redirectTo = "/"
		}
		sess.Delete(SessionRedirectKey)
	}

	// User authenticated.
	sess.Set("user", *user)
	if err := SaveSession(sess, ctx); err != nil {
		ctx.Error(fmt.Errorf("completeUserAuth error: session write/read went bad! Error: %#v", err))
		return
	}
	fmt.Printf("User authenticate via discord! %q %s#%s", u.UserID, u.Name, u.RawData["discriminator"])

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

func (o *Oauth) SaveUser(ctx *gin.Context, user *data.User) error {

	if err := o.Web.Data.Engine.Save(&user).Error; err != nil {
		return err
	}

	sess := sessions.Default(ctx)
	sess.Set("user", user)
	if err := SaveSession(sess, ctx); err != nil {
		return err
	}

	return nil
}

func (o *Oauth) GetUser(ctx *gin.Context) (data.User, bool) {
	sess := sessions.Default(ctx)
	user, ok := sess.Get("user").(data.User)
	return user, ok
}

func (o *Oauth) GetUserOrRedirect(ctx *gin.Context) data.User {
	sess := sessions.Default(ctx)
	user, ok := o.GetUser(ctx)
	if !ok || user.Agreed {
		if redirectTo := sess.Get(SessionRedirectKey); redirectTo != nil {
			sess.Delete(SessionRedirectKey)
			if err := SaveSession(sess, ctx); err != nil {
				ctx.Error(err)
				return data.User{}
			}

			if to, ok := redirectTo.(string); ok && to != "" {
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

func (o *Oauth) CreateOrGetUser(ctx *gin.Context, u goth.User) (*data.User, error) {
	userId, err := strconv.ParseUint(u.UserID, 10, 64)
	if err != nil {
		return nil, err
	}

	user := &data.User{
		ID:    userId,
		Email: u.Email,
	}

	// engine create.
	engine := o.Web.Data.Engine
	engine.Where(user).First(user)

	user.Name = u.Name
	user.Discriminator = u.RawData["discriminator"].(string)
	user.SessionToken = u.AccessToken
	user.SessionTokenSecret = u.AccessTokenSecret
	user.RefreshToken = u.RefreshToken
	user.Expiry = u.ExpiresAt

	if user.CreatedAt.IsZero() {
		return user, engine.Create(user).Error
	} else {
		return user, engine.Save(user).Error
	}
}

func (o *Oauth) refreshToken(u goth.User) error {
	if time.Now().After(u.ExpiresAt) {
		token, err := discordProvider.RefreshToken(u.RefreshToken)
		if err != nil {
			return err
		}

		u.RefreshToken = token.RefreshToken
		u.AccessToken = token.AccessToken
		u.ExpiresAt = token.Expiry
	}

	return nil
}

func getAuthSessionFromSession(ctx *gin.Context) (goth.Session, error) {
	sess := sessions.Default(ctx)
	str, ok := sess.Get(SessionAuthKey).(string)
	if !ok {
		return nil, errors.New("invalid session auth data")
	}
	return discordProvider.UnmarshalSession(str)
}

func getAuthUserFromSession(ctx *gin.Context) (goth.User, error) {
	oauthSess, err := getAuthSessionFromSession(ctx)
	if err != nil {
		return goth.User{}, err
	}

	return discordProvider.FetchUser(oauthSess)
}
