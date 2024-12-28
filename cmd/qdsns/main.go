package main

import (
	"fmt"
	sns "github.com/aldrinleal/qdsns"
	"github.com/gin-gonic/gin"
	"github.com/joomcode/errorx"
	"github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
	"net/http"
	"os"
	"strings"
)

func getPort() string {
	listener := ":8000"

	if newListenerPort, exists := os.LookupEnv("PORT"); exists {
		listener = ":" + newListenerPort
	}

	return listener
}

var log = logrus.New()

func main() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	e := gin.Default()
	e.Use(ginlogrus.Logger(log), gin.Recovery())

	genericHandler := func(c *gin.Context) {
		log.Infof("Request for %s", c.Request.URL)
		for k, v := range c.Request.Header {
			log.Infof(" %s: %s", k, strings.Join(v, "; "))
		}

		c.JSON(200, gin.H{
			"status": "ok",
		})
	}

	e.GET("/", genericHandler)
	e.GET("/health", genericHandler)
	e.GET("/healthcheck", genericHandler)

	e.Any("/any/*any", genericHandler)

	e.GET("/sns/:id", func(c *gin.Context) {
		if "GET" == c.Request.Method {
			id := c.Param("id")

			handlerUrl, _ := c.Request.URL.Parse("/sns/" + id)

			c.JSON(200, gin.H{
				"status": "ok",
				"id":     id,
				"url":    handlerUrl.String(),
			})
		}
	})

	e.POST("/sns/:id", func(c *gin.Context) {
		log.Infof("SNS")

		reportFailure := func(c *gin.Context, statusCode int, err error) {
			log.Warnf("Oops: %s", err)

			c.JSON(statusCode, gin.H{
				"code":    fmt.Sprintf("%03d", statusCode),
				"status":  "failure",
				"message": err.Error(),
			})
		}

		snsMessage := &sns.Notification{}

		if err := c.BindJSON(snsMessage); nil != err {
			reportFailure(c, 500, errorx.Decorate(err, "while binding snsMessage"))

			return
		}

		if err := snsMessage.VerifySignature(); nil != err {
			reportFailure(c, 400, errorx.Decorate(err, "while validating signature"))

			return
		}

		if "SubscriptionConfirmation" == snsMessage.Type {
			log.Infof("Confirming Subscription to topic %s", snsMessage.TopicArn)

			_, err := snsMessage.Subscribe()

			if nil != err {
				reportFailure(c, 500, errorx.Decorate(err, "confirming subscription"))

				return
			}
		} else if "Notification" == snsMessage.Type {
			log.Infof("Body: %s", snsMessage.Message)
		}

	})

	listener := getPort()

	log.Infof("Going to listen on %s", listener)

	log.Fatalf("Oops: %s", http.ListenAndServe(listener, e))
}
