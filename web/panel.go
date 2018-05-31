package web

import "github.com/gin-gonic/gin"

type Config struct {
	StaticPath string `json:"static_path"`
}

type Panel struct {
	engine *gin.Engine
}

func New(config *Config) (ret *Panel, err error) {
	ret = &Panel{
		engine: gin.Default(),
	}

	ret.engine.Static("/static/", config.StaticPath )

	return
}