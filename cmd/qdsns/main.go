package main

import (
	"fmt"
	sns "github.com/aldrinleal/qdsns"
	"github.com/gin-gonic/gin"
	"github.com/joomcode/errorx"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func getPort() string {
	listener := ":8000"

	if newListenerPort, exists := os.LookupEnv("PORT"); exists {
		listener = ":" + newListenerPort
	}
	return listener
}

func main() {
	e := gin.Default()

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

	log.Fatalf("Oops: %s", http.ListenAndServe(listener, e))
}
