package main

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Response string `json:"response"`
	Message  string `json:"message"`
	Code     int    `json:"status"`
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, Response{Response: "Main Page", Message: "OK", Code: 200})
	})

	r.Run(":8080")
}
