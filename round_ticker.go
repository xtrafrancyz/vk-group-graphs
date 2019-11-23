package main

import (
	"log"
	"time"
)

type Ticker struct {
	Name      string
	Precision time.Duration
	Callback  func()
}

func (t Ticker) Start() {
	now := time.Now()
	startAt := now.Round(t.Precision)
	if startAt.Before(now) {
		startAt = startAt.Add(t.Precision)
	}
	sleepDuration := startAt.Sub(now)

	log.Printf("[%s] First iteration will be after %s", t.Name, sleepDuration)

	time.AfterFunc(sleepDuration, func() {
		t.Callback()
		ticker := time.NewTicker(t.Precision)
		go func() {
			for range ticker.C {
				t.Callback()
			}
		}()
	})
}
