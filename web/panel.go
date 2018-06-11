package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/SelfBotBot/selfbot/config"
	"github.com/SelfBotBot/selfbot/data"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

type Panel struct {
	Gin       *gin.Engine
	Config    *config.Config
	Redis     *redis.Pool
	Data      *data.Handler
	PlaySound func(guildId, sound string) bool
}

const (
	SessionAuthKey     = "auth.session"
	SessionRedirectKey = "auth.redirectTo"
)

// Handler an interface for sections of the site that handle.
type Handler interface {
	RegisterHandlers() error
}

func New(config *config.Config) (ret *Panel, err error) {
	ret = &Panel{
		Gin:    gin.Default(),
		Config: config,
	}

	if err = ret.AddPreMiddleware(); err != nil {
		return
	}

	if err = ret.AddPostMiddleware(); err != nil {
		return
	}

	oauth := Oauth{Web: ret}
	oauth.RegisterHandlers()

	alexaMeme := &AlexaMeme{Web: ret}
	alexaMeme.RegisterHandlers()

	return
}

func (panel *Panel) RunAutoTLS() error {
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(panel.Config.Web.DomainNames[0:]...),
	}
	dir := cacheDir()
	fmt.Println("Using cache: ", dir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Printf("warning: autocert.NewListener not using a cache: %v", err)
	} else {
		m.Cache = autocert.DirCache(dir)
	}
	go http.ListenAndServe(":http", m.HTTPHandler(nil))
	return autotls.RunWithManager(panel.Gin, *m)
}

func SaveSession(sesh sessions.Session, ctx *gin.Context) error {
	if err := sesh.Save(); err != nil {
		ctx.Error(err)
		return err
	}
	return nil
}

func cacheDir() string {
	const base = "golang-autocert"
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir(), "Library", "Caches", base)
	case "windows":
		for _, ev := range []string{"APPDATA", "CSIDL_APPDATA", "TEMP", "TMP"} {
			if v := os.Getenv(ev); v != "" {
				return filepath.Join(v, base)
			}
		}
		// Worst case:
		return filepath.Join(homeDir(), base)
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, base)
	}
	return filepath.Join(homeDir(), ".cache", base)
}

func homeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return "/"
}
