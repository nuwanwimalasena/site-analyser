package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	// Load HTML templates
	router.LoadHTMLGlob("templates/*")
	// Register routes
	router.GET("/", FormHandler)
	router.POST("/results", ResultsHandler)

	router.Run(":8080")
}
