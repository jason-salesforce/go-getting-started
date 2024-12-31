package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/exp/maps"
	// "net/url"
	// "strings"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

type HerokuSystem struct {
	Uri            string         `json:"uri"`
	DownstreamApps []HerokuSystem `json:"downstream_apps"`
}

type Payload struct {
	System       HerokuSystem      `json:"system"`
	SystemChecks map[string]string `json:"system_checks,omitempty"`
}

type HealthCheckResponse struct {
	Status       string            `json:"status"`
	StatusCode   int               `json:"status_code"`
	SystemChecks map[string]string `json:"system_checks"`
}

func main() {
	port := os.Getenv("PORT")
	domain := "http://localhost:" + port + "/health"

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.POST("/health", func(c *gin.Context) {
		if os.Getenv("DEMO_FAIL") != "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "down",
			})
		}
		var payload Payload

		if err := c.ShouldBindJSON(&payload); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if payload.SystemChecks == nil {
			payload.SystemChecks = make(map[string]string)
			payload.SystemChecks["init -> http://localhost:5000/health"] = "up"
		} else {
			payload.SystemChecks[c.Request.Referer()+" -> "+domain] = "up"
		}

		for _, app := range payload.System.DownstreamApps {
			newPayload := Payload{}
			newPayload.SystemChecks = payload.SystemChecks
			newPayload.System = app

			nextJson, _ := json.Marshal(newPayload)

			client := http.Client{}
			req, _ := http.NewRequest("POST", app.Uri, bytes.NewBuffer(nextJson))
			req.Header.Set("Referer", domain)

			resp, _ := client.Do(req)

			r := HealthCheckResponse{}
			_ = json.NewDecoder(resp.Body).Decode(&r)
			maps.Copy(payload.SystemChecks, r.SystemChecks)

			fmt.Println(r)
		}

		c.JSON(200, gin.H{
			"system_checks": payload.SystemChecks,
		})

	})

	router.Run(":" + port)
}
