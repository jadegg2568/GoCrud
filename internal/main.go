package main

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func main() {
	r := gin.Default()

	connectAddr := ""

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, Response{Code: "OK", Message: "Main Page"})
	})

	r.Run(":8080")
}
