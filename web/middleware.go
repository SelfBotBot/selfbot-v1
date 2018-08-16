package web

import (
	"regexp"
	"strings"
	"time"

	"github.com/SelfBotBot/selfbot/data"
	"github.com/SilverCory/gin-redisgo-cooldowns"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	redstore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/unrolled/secure"
	"github.com/utrack/gin-csrf"
)

var allowedInRegisterRegex = regexp.MustCompile(`(?i)^(/(logout|register|tos|privacy|((js|css|img|auth)/*.)))|/$`)

const CSP = `
default-src 'self';
img-src 'self' https://cdnjs.cloudflare.com/ https://placekitten.com/ https://cdn.discordapp.com/;
script-src 'self' https://cdnjs.cloudflare.com/ajax/libs/cookieconsent2/3.0.3/cookieconsent.min.js 'sha256-SplWdsqEBp8LjzZSKYaEfDXhXSi0/oXXxAnQSYREAuI=';
style-src 'self' https://cdnjs.cloudflare.com/ajax/libs/cookieconsent2/3.0.3/cookieconsent.min.css 'unsafe-inline';
`

type Middleware struct {
	web *Panel
}

var m *Middleware

func (panel *Panel) AddPreMiddleware() (err error) {
	m = &Middleware{panel}

	if err = m.setupSessions(); err != nil {
		return
	}

	return
}

func (panel *Panel) AddPostMiddleware() (err error) {
	m.setupCors()
	m.setupCsrf()
	m.setupSecurity()
	m.setupRegisterRedirect()
	m.setupIPCooldowns()

	return
}

func (m *Middleware) setupCsrf() {
	m.web.Gin.Use(csrf.Middleware(csrf.Options{
		Secret: m.web.Config.Web.CSRFSecret,
		ErrorFunc: func(c *gin.Context) {

			if c.Request.URL.Path == "/CSPReport" || strings.HasPrefix(c.Request.URL.Path, "/alexamemes/") {
				return
			}

			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))
}

func (m *Middleware) setupSessions() (err error) {
	conf := m.web.Config.Redis
	if !(conf.Enabled) {
		store := cookie.NewStore([]byte("dankest_selfbot_ever"))
		m.web.Gin.Use(sessions.Sessions("selfbot_sessions", store))
		return
	}

	//TODO this is here, idk why

	m.web.Redis = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dialWithDB(conf.Network, conf.Address, conf.Password, conf.Database)
		},
	}

	store, err := redstore.NewStoreWithPool(m.web.Redis, []byte("dankest_selfbot_ever"))
	if err != nil {
		return
	}

	if redstore.SetKeyPrefix(store, "selfbot.sessions.sesh:"); err != nil {
		panic(err)
	}

	store.Options(sessions.Options{
		Secure:   true,
		MaxAge:   int(((24 + time.Hour) * 7).Seconds()),
		HttpOnly: true,
		Domain:   "sb.cory.red",
	})

	m.web.Gin.Use(sessions.Sessions("selfbot_sessions", store))
	return nil
}

func (m *Middleware) setupIPCooldowns() {
	m.web.Gin.Use(gin_redisgo_cooldowns.NewRateLimit(m.web.Redis, "selfbot.cooldown.general.ip:", 100, time.Second*5, nil))
}

func (m *Middleware) setupCors() {
	if gin.IsDebugging() {
		return
	}

	m.web.Gin.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://cdnjs.cloudflare.com", "https://placekitten.com", "https://sb.cory.red"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
}

func (m *Middleware) setupSecurity() {
	sec := secure.New(secure.Options{
		AllowedHosts:            []string{"sb.cory.red"},
		SSLRedirect:             false,
		SSLTemporaryRedirect:    false,
		SSLHost:                 "sb.cory.red",
		STSSeconds:              86400,
		STSIncludeSubdomains:    true,
		STSPreload:              true,
		ForceSTSHeader:          true,
		FrameDeny:               true,
		CustomFrameOptionsValue: "SAMEORIGIN",
		ContentTypeNosniff:      true,
		BrowserXssFilter:        true,
		ContentSecurityPolicy:   CSP,
		HostsProxyHeaders:       []string{"X-Forwarded-For"},

		IsDevelopment: gin.IsDebugging(),
	})

	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			err := sec.Process(c.Writer, c.Request)

			// If there was an error, do not continue.
			if err != nil {
				c.Abort()
				return
			}

			// Avoid header rewrite if response is a redirection.
			if status := c.Writer.Status(); status > 300 && status < 399 {
				c.Abort()
			}
		}
	}()

	m.web.Gin.Use(secureFunc)
}

func (m *Middleware) setupRegisterRedirect() {
	m.web.Gin.Use(func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if allowedInRegisterRegex.MatchString(path) {
			ctx.Next()
			return
		}

		// TODO sesh
		sess := sessions.Default(ctx)
		user, ok := sess.Get("user").(data.User)
		if !ok {
			ctx.Next()
			return
		}

		if !user.Agreed {
			sess.Set(SessionRedirectKey, path)
			ctx.Redirect(302, "/register")
			return
		}

		ctx.Next()
	})
}

func dialWithDB(network, address, password, DB string) (redis.Conn, error) {
	c, err := dial(network, address, password)
	if err != nil {
		return nil, err
	}
	if _, err := c.Do("SELECT", DB); err != nil {
		c.Close()
		return nil, err
	}
	return c, err
}

func dial(network, address, password string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}
