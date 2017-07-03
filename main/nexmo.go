package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/njern/gonexmo.v1"
)

func setupNexmo(a *App) {

	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	var config map[string]interface{}

	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error: ", err)
	}

	apiKey, _ := config["nexmo_api_key"].(string)
	secret, _ := config["nexmo_secret"].(string)
	nexmoClient, _ := nexmo.NewClientFromAPI(apiKey, secret)
	a.nexmoClient = nexmoClient
}

func sendMessage(nexmoClient *nexmo.Client, to, text string) {
	message := &nexmo.SMSMessage{
		From: "12016219524",
		To:   to,
		Type: nexmo.Text,
		Text: text,
	}
	nexmoClient.SMS.Send(message)
}
