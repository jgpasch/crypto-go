package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

type sub struct {
	ID           int    `json:"id"`
	Token        string `json:"token"`
	FullName     string
	Percent      float64 `json:"percent"`
	MinVal       float64 `json:"minVal"`
	MaxVal       float64 `json:"maxVal"`
	MinMaxChange float64 `json:"minMaxChange"`
	Owner        string  `json:"owner"`
	Active       bool    `json:"active"`
}

func (s *sub) getSubByToken(db *sql.DB) error {
	return db.QueryRow("SELECT token, percent, minval, maxval, minmaxchange, active FROM subs WHERE token=$1 AND owner=$2",
		s.Token, s.Owner).Scan(&s.Token, &s.Percent, &s.MinVal, &s.MaxVal, &s.MinMaxChange, &s.Active)
}

func (s *sub) getSub(db *sql.DB) error {
	return db.QueryRow("SELECT token, percent, minval, maxval, minmaxchange, active FROM subs WHERE id=$1 AND owner=$2",
		s.ID, s.Owner).Scan(&s.Token, &s.Percent, &s.MinVal, &s.MaxVal, &s.MinMaxChange, &s.Active)
}

func (s *sub) updateSub(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE subs SET token=$1, percent=$2, minval=$3, maxval=$4, minmaxchange=$5, active=$6 WHERE id=$7",
			s.Token, s.Percent, s.MinVal, s.MaxVal, s.MinMaxChange, s.Active, s.ID)

	return err
}

func (s *sub) toggleActive(db *sql.DB) error {
	var temp sub
	temp.Owner = s.Owner
	temp.ID = s.ID
	err := temp.getSub(db)
	if err != nil {
		return err
	}
	newActive := !temp.Active

	_, err2 :=
		db.Exec("UPDATE subs SET active=$1 WHERE id=$2 AND owner=$3",
			newActive, s.ID, s.Owner)
	s.Active = newActive

	return err2
}

func (s *sub) deleteSub(db *sql.DB) error {
	_, err :=
		db.Exec("DELETE FROM subs WHERE id=$1", s.ID)

	return err
}

func (s *sub) createSub(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO subs(token, percent, minval, maxval, minmaxchange, owner) VALUES($1, $2, $3, $4, $5, $6) RETURNING id",
		s.Token, s.Percent, s.MinVal, s.MaxVal, s.MinMaxChange, s.Owner).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

func getAllSubs(db *sql.DB) ([]sub, error) {
	rows, err := db.Query("SELECT id, active FROM subs")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	subs := []sub{}

	for rows.Next() {
		var s sub
		if err := rows.Scan(&s.ID, &s.Active); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	return subs, nil
}

func getAllSubsByOwner(db *sql.DB, owner string) ([]sub, error) {
	rows, err := db.Query(
		"SELECT id, token, percent, minval, maxval, minmaxchange, owner, active FROM subs WHERE owner=$1", owner)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	subs := []sub{}

	for rows.Next() {
		var s sub
		if err := rows.Scan(&s.ID, &s.Token, &s.Percent, &s.MinVal, &s.MaxVal, &s.MinMaxChange, &s.Owner, &s.Active); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	return subs, nil
}

func (s *sub) doEvery(a *App) {

	// add this watch to the map
	ticker := time.NewTicker(4 * time.Second)
	myMap[s.ID] = watcher{WatchEvent: ticker}

	fmt.Println(s.ID)
	err := a.DB.QueryRow("SELECT token, owner FROM subs where id=$1", s.ID).Scan(&s.Token, &s.Owner)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("no rows")
	case err != nil:
		fmt.Println(err)
	default:
	}

	var number string
	err2 := a.DB.QueryRow("SELECT number FROM users where email=$1", s.Owner).Scan(&number)
	switch {
	case err2 == sql.ErrNoRows:
		fmt.Println("no user with that email")
	case err2 != nil:
		fmt.Println(err)
	default:
	}

	go func() {
		var lastPrice float64
		for t := range myMap[s.ID].WatchEvent.C {
			fmt.Println("Tick at: ", t, " --  ", s.ID)
			// reach out to coin market cap api
			var m = make([]map[string]interface{}, 0)
			getJSON("https://api.coinmarketcap.com/v1/ticker/"+symbols[s.Token]+"/", &m)
			// fmt.Println(m[0])
			tmp, _ := strconv.ParseFloat(m[0]["price_usd"].(string), 64)
			fmt.Println(lastPrice / tmp)

			if lastPrice == 0.0 {
				// first time starting watch, set init price

				fmt.Println("setting price for first time:" + m[0]["price_usd"].(string))
				lastPrice = tmp
			} else if math.Abs(lastPrice/tmp) > (1 + .1) {
				fmt.Println("top floor hit, new price': " + m[0]["price_usd"].(string))
				lastPrice = tmp
			} else if math.Abs(lastPrice/tmp) < (1 - .1) {
				fmt.Println("bottom floor hit")
				lastPrice = tmp
			} else {
				fmt.Println("difference not enough, move on")
			}

			// lastPrice, _ = strconv.ParseFloat(m[0]["price_usd"].(string), 64)

			// fmt.Println(s.Token + ":  " + m[0]["price_usd"].(string))
			// sendMessage(a.nexmoClient, number, s.Token+":  "+strconv.FormatFloat(lastPrice, 'f', -1, 64))

		}
	}()
}

func (s *sub) stopWatch() {
	myMap[s.ID].WatchEvent.Stop()
	delete(myMap, s.ID)
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string, target interface{}) error {
	res, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b := json.NewDecoder(res.Body).Decode(target)
	return b
}
