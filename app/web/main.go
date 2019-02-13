package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("resource/template/**/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", gin.H{})
	})
	r.Run()
}
