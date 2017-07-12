package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	a := App{}
	a.Initialize("john", "new_sub_db")

	setupNexmo(&a)

	a.startWatchEvents()

	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	var config map[string]interface{}

	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error: ", err)
	} else {
		a.Run(config["port"].(string))
	}
}
