package main

import (
	"log"
	"strconv"
	"sync"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
)

type Counter struct {
	storage *Storage
	lock    *sync.Mutex
	users   map[int]int
}

func NewCounter(storage *Storage) *Counter {
	c := Counter{
		storage: storage,
		lock:    &sync.Mutex{},
		users:   make(map[int]int),
	}
	Ticker{
		Name:      "messages",
		Precision: time.Hour,
		Callback:  c.save,
	}.Start()
	return &c
}

func (c *Counter) OnMessageReply(from, to int) {
	c.lock.Lock()
	c.users[from]++
	c.lock.Unlock()
}

func (c *Counter) save() {
	log.Println("Save " + string(len(c.users)))

	c.lock.Lock()

	for from, messages := range c.users {
		tags := map[string]string{
			"agent": strconv.Itoa(from),
		}
		fields := map[string]interface{}{
			"messages": messages,
		}
		point, err := influx.NewPoint("messages", tags, fields, time.Now())
		if err != nil {
			log.Println("Could not create point: ", err.Error())
		} else {
			c.storage.PointsChannel <- point
		}
	}
	c.users = make(map[int]int)

	c.lock.Unlock()
}
