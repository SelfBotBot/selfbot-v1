package web

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-contrib/static"

	"github.com/SelfBotBot/selfbot/data"
	"github.com/gin-contrib/sessions"

	"github.com/SelfBotBot/selfbot/config"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
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
		Gin:    gin.New(),
		Config: config,
	}

	if f, err := os.Stat(config.Web.LogDirectory); os.IsNotExist(err) || !f.IsDir() {
		if err := os.MkdirAll(config.Web.LogDirectory, 0644); err != nil {
			return nil, err
		}
	}

	f, err := os.Create(config.Web.LogDirectory + "/gin.log")
	if err != nil {
		return ret, err
	}

	errF, err := os.Create(config.Web.LogDirectory + "/gin_err.log")
	if err != nil {
		return ret, err
	}
	ret.Gin.Use(gin.LoggerWithWriter(io.MultiWriter(f)))
	ret.Gin.Use(gin.RecoveryWithWriter(io.MultiWriter(errF, os.Stderr)))

	// Load the HTML templates
	// Templating
	ret.Gin.SetFuncMap(template.FuncMap{
		"comments": func(s string) template.HTML { return template.HTML(s) },
		"ASCII":    GetAscii,
	})
	ret.Gin.LoadHTMLGlob(config.Web.TemplateGlob)

	// Static files to load
	ret.Gin.Use(static.Serve("/", static.LocalFile(ret.Config.Web.StaticFilePath, false)))

	if err = ret.AddPreMiddleware(); err != nil {
		return
	}

	if err = ret.AddPostMiddleware(); err != nil {
		return
	}

	oauth := Oauth{Web: ret}
	oauth.RegisterHandlers()

	pages := &Pages{Web: ret}
	pages.RegisterHandlers()

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
	go http.ListenAndServe(panel.Config.Web.ListenAddress+":80", m.HTTPHandler(nil))
	return runWithManager(panel.Gin, *m, panel.Config.Web.ListenAddress)
}

func (panel *Panel) GetUser(ctx *gin.Context) (data.User, bool) {
	sess := sessions.Default(ctx)
	user, ok := sess.Get("user").(data.User)
	return user, ok
}

func runWithManager(r http.Handler, m autocert.Manager, address string) error {
	s := &http.Server{
		Addr:              address + ":443",
		TLSConfig:         &tls.Config{GetCertificate: m.GetCertificate},
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      15 * time.Second,
		MaxHeaderBytes:    2048,
	}

	return s.ListenAndServeTLS("", "")
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
