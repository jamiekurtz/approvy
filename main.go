package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sfreiberg/gotwilio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

var config, secrets *viper.Viper
var db *gorm.DB

func init() {
	initConfig()
	initDb()

	if config.GetString("DEBUG_LOG_LEVEL") == "yes" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithWriter(os.Stdout, "/status"))

	r.GET("/", getIndex)
	r.GET("/status", getStatus)
	r.GET("/requests/:id", getApprovalRequestsHandler)
	r.POST("/requests", postApprovalRequestHandler)
	r.POST("/requests/:id/responses", postApprovalResponseHandler)

	log.Info("Starting approvy server on port 3000.")
	log.Info("Browse to http://localhost:3000 to see the default home page.")

	r.Run(":3000")
}

func initConfig() {
	config = viper.New()
	secrets = viper.New()
	loadConfigFile(config, "config/config")
	loadConfigFile(secrets, "config/secrets")
}

func initDb() {
	os.Remove("./approvy_v1.db")

	var err error

	db, err = gorm.Open("sqlite3", "./approvy_v1.db")
	if err != nil {
		log.WithError(err).Fatal("Error initializing database")
	}

	db.AutoMigrate(&Request{}, &Response{})
}

func loadConfigFile(v *viper.Viper, filename string) {
	v.SetConfigName(filename)
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		log.WithError(err).Fatal("Error loading config files")
	}

	v.AutomaticEnv()
}

func getIndex(c *gin.Context) {
	c.String(200, "Approvy!")
}

func getStatus(c *gin.Context) {
	c.Status(200)
}

func getApprovalRequestsHandler(c *gin.Context) {
	id := c.Param("id")

	request := Request{}
	found := db.Preload("Responses").First(&request, id).Error != gorm.ErrRecordNotFound

	if !found {
		c.Status(404)
		return
	}

	c.JSON(200, request)
}

func postApprovalRequestHandler(c *gin.Context) {
	to := c.PostForm("to")
	message := c.PostForm("message")
	from := c.PostForm("from")

	expiresAt := time.Now().Add(time.Hour)
	request := Request{From: from, To: to, Message: message, ExpiresAt: expiresAt}
	db.Create(&request)

	c.String(200, request.IDstr())

	twilioEnabled := config.GetString("TWILIO_ENABLED")
	if twilioEnabled == "true" {
		sendApprovalRequest(from, to, message)
	}
}

func postApprovalResponseHandler(c *gin.Context) {
	id := c.Param("id")

	request := Request{}
	found := db.Preload("Responses").First(&request, id).Error != gorm.ErrRecordNotFound

	if !found {
		c.Status(404)
		return
	}

	approvedStr := c.PostForm("approved")
	approved := approvedStr == "true"
	response := Response{RequestID: request.ID, Approved: approved}
	db.Create(&response)

	if approved {
		for _, r := range request.Responses {
			approved = approved && r.Approved
		}
	}

	request.Approved = approved
	db.Save(&request)

	c.Status(200)
}

func sendApprovalRequest(from string, to string, subject string) {
	accountSid := config.GetString("TWILIO_ACCOUNT_SID")
	authToken := secrets.GetString("TWILIO_AUTH_TOKEN")
	from_number := config.GetString("TWILIO_FROM")

	twilio := gotwilio.NewTwilioClient(accountSid, authToken)

	to_number := getApproverNumber(to)
	message := fmt.Sprintf("Approval request from %s regarding: %s", from, subject)
	_, ex, err := twilio.SendSMS(from_number, to_number, message, "", "")
	if err != nil {
		log.WithError(err).Error("Error sending message with Twilio")
	}
	if ex != nil {
		log.Errorf("Exception sending message with Twilio: %s", ex)
	}
}

func getApproverNumber(approver string) string {
	key := "approvers." + approver + ".sms"
	number := config.GetString(key)
	if number != "" {
		return number
	}

	panic(fmt.Sprintf("Unknown approver '%s'", approver))
}
