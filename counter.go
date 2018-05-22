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
	return &Counter{
		storage: storage,
		lock:    &sync.Mutex{},
		users:   make(map[int]int),
	}
}

func (c *Counter) OnMessageSent(from, to int) {
	c.lock.Lock()
	c.users[from]++
	c.lock.Unlock()
}

func (c *Counter) Save() {
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

func (c *Counter) ScheduleSave() {
	now := time.Now()
	startOfNextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	startOfNextHour = startOfNextHour.Add(time.Hour)
	sleepDuration := startOfNextHour.Sub(now)

	log.Printf("Save will be proceed in %d minutes", int(sleepDuration.Minutes()))

	time.AfterFunc(sleepDuration, func() {
		c.Save()
		ticker := time.NewTicker(time.Hour)
		go func() {
			for range ticker.C {
				c.Save()
			}
		}()
	})

	time.NewTimer(time.Second)
}
