package main

import (
    "fmt"
    "net/http"
    "log"
    "github.com/gorilla/mux"
    "github.com/sfreiberg/gotwilio"
    "github.com/spf13/viper"
)

var config, secrets *viper.Viper

func init() {
    initConfig();
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/", index)
    router.HandleFunc("/requests", approvalRequestHandler).Methods("POST")

    fmt.Println("Starting approvy server on port 3000.")
    fmt.Println("Browse to http://localhost:3000 to see the default home page.")

    log.Fatal(http.ListenAndServe(":3000", router))
}

func initConfig() {
    config = viper.New()
    secrets = viper.New()
    loadConfigFile(config, "config/config")
    loadConfigFile(secrets, "config/secrets")
}

func loadConfigFile(v *viper.Viper, filename string) {
    v.SetConfigName(filename)
    v.AddConfigPath(".")
    err := v.ReadInConfig()
    if err != nil {
        panic(fmt.Errorf("Fatal error config file: %s \n", err))
    }

    v.AutomaticEnv()
}

func index(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Approvy!\n"))
}

func approvalRequestHandler(w http.ResponseWriter, r *http.Request) {
    to := r.FormValue("to")
    subject := r.FormValue("subject")
    from := r.FormValue("from")

    sendApprovalRequest(from, to, subject)

    w.Write([]byte("message sent"))
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
        fmt.Println("Error occured...")
        fmt.Println("Error sending message with Twilio: %s", err)
    }
    if ex != nil {
        fmt.Println("Exception occured...")
        fmt.Println("Exception sending message wiht Twilio: %s", ex)
    }
}

func getApproverNumber(approver string) (string) {
    key := "approvers." + approver + ".sms"
    number := config.GetString(key)
    if number != "" {
        return number
    }

    panic(fmt.Sprintf("Unknown approver '%s'", approver))
}

