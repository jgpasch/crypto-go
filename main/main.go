package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var symbols map[string]string

func main() {
	a := App{}
	a.Initialize("john", "new_sub_db")

	setupNexmo(&a)

	a.startWatchEvents()

	// get map of symbols to coin names
	syms, _ := os.Open("../symbols.json")
	decoder2 := json.NewDecoder(syms)

	err2 := decoder2.Decode(&symbols)
	if err2 != nil {
		fmt.Println("error: ", err2)
	} else {
		fmt.Println("got them symbols")
	}

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
