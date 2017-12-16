package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sfreiberg/gotwilio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
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
	router := mux.NewRouter()
	router.HandleFunc("/", index)
	router.HandleFunc("/status", status)
	router.HandleFunc("/requests/{id}", getApprovalRequestsHandler).Methods("GET")
	router.HandleFunc("/requests", postApprovalRequestHandler).Methods("POST")
	router.HandleFunc("/requests/{id}/responses", postApprovalResponseHandler).Methods("POST")

	log.Info("Starting approvy server on port 3000.")
	log.Info("Browse to http://localhost:3000 to see the default home page.")

	log.Fatal(http.ListenAndServe(":3000", router))
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

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Approvy!\n"))

	request := Request{}
	db.Last(&request)
	w.Write([]byte(request.Message))
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func getApprovalRequestsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	request := Request{}
	found := db.Preload("Responses").First(&request, id).Error != gorm.ErrRecordNotFound
	b, err := json.Marshal(request)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if !found {
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func postApprovalRequestHandler(w http.ResponseWriter, r *http.Request) {
	to := r.FormValue("to")
	message := r.FormValue("message")
	from := r.FormValue("from")

	expiresAt := time.Now().Add(time.Hour)
	request := Request{From: from, To: to, Message: message, ExpiresAt: expiresAt}
	db.Create(&request)

	w.Write([]byte(request.IDstr()))

	twilioEnabled := config.GetString("TWILIO_ENABLED")
	if twilioEnabled == "yes" {
		sendApprovalRequest(from, to, message)
	}
}

func postApprovalResponseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	request := Request{}
	found := db.Preload("Responses").First(&request, id).Error != gorm.ErrRecordNotFound

	if !found {
		w.WriteHeader(404)
		return
	}

	approvedStr := r.FormValue("approved")
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
