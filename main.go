package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"os"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.GET("/health", func(c *gin.Context) {
		if os.Getenv("DEMO_FAIL") != "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "down",
			})
		}
		paths := c.Query("paths")
		var resp *http.Response
		var err error

		if paths != "" {
			fmt.Println("path found")
			str, _ := url.QueryUnescape(paths)
			result := strings.Split(str, ",")

			if len(result) > 1 {
				nextPath := url.QueryEscape(strings.Join(result[1:], ","))
		 	  resp, err = http.Get(result[0] + "?paths=" + nextPath)
			} else {
				resp, err = http.Get(result[0])
			}

			if err != nil || resp.StatusCode != 200 {
				c.JSON(resp.StatusCode, gin.H{
					"status": "down",
				})

			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": "up",
				})
			}
		} else {
				fmt.Println("no path found")
				c.JSON(http.StatusOK, gin.H{
					"status": "up",
				})
		}
	})

	router.Run(":" + port)
}
