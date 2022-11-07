package main

import (
	"strconv"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Counter struct {
	storage *Storage
	lock    sync.Mutex
	users   map[int]int
}

func NewCounter(storage *Storage) *Counter {
	c := Counter{
		storage: storage,
		users:   make(map[int]int),
	}
	Ticker{
		Name:      "messages",
		Precision: time.Hour,
		Callback:  c.save,
	}.Start()
	return &c
}

func (c *Counter) OnMessageOut(from, to int) {
	c.lock.Lock()
	c.users[from]++
	c.lock.Unlock()
}

func (c *Counter) save() {
	c.lock.Lock()

	for from, messages := range c.users {
		c.storage.AddPoint(influxdb2.NewPoint(
			"messages",
			map[string]string{
				"agent": strconv.Itoa(from),
			},
			map[string]any{
				"messages": messages,
			},
			time.Now(),
		))
	}
	c.users = make(map[int]int)

	c.lock.Unlock()
}
