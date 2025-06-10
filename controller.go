package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func FormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "form.html", gin.H{})
}

func ResultsHandler(c *gin.Context) {
	url := c.PostForm("url")
	if !ValidateURL(url) {
		c.HTML(http.StatusBadRequest, "form.html", gin.H{
			"Error": "Invalid URL, Please enter a valid URL",
		})
		return
	}
	r, e := ReadPageContent(url)
	if e != nil {
		c.HTML(http.StatusBadRequest, "form.html", gin.H{
			"Error": e.Error(),
		})
		return
	}
	c.HTML(http.StatusOK, "results.html", gin.H{
		"Url":         url,
		"HTMLVersion": r.HTMLVersion,
		"Title":       r.Title,
		"Headings":    r.Headings,
		"Links":       r.Links,
		"LoginForm":   r.LoginForm,
	})
}
