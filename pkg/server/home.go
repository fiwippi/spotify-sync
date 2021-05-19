package server

import (
	"embed"
	"github.com/gin-gonic/gin"
	"net/http"
)

//go:embed static/*
var efs embed.FS

func favicon(c *gin.Context) {
	c.FileFromFS("static/favicon.png", http.FS(efs))
}

func admin(c *gin.Context) {
	c.HTML(200, "admin.tmpl", nil)
}
