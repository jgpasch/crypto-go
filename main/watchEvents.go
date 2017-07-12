package main

import (
	"fmt"
	"time"
)

type watcher struct {
	WatchEvent *time.Ticker
}

var myMap = map[int]watcher{}

func (a *App) startWatchEvents() {
	// get all subscriptions, if any of them are active, then start the watch
	subs, err := getAllSubs(a.DB)
	if err != nil {
		fmt.Println(err.Error())
	}
	for i := range subs {
		if subs[i].Active {
			subs[i].doEvery()
		}
	}
}
