package main

import (
	"log"
	"time"

	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/xtrafrancyz/vk-group-graphs/vkapi"
)

type Unread struct {
	storage *Storage
	api     *vkapi.Api
	groupId string
}

func NewUnread(storage *Storage, api *vkapi.Api, vkGroupId string) *Unread {
	u := Unread{
		storage: storage,
		api:     api,
		groupId: vkGroupId,
	}
	Ticker{
		Name:      "unread",
		Precision: 5 * time.Minute,
		Callback:  u.gather,
	}.Start()
	return &u
}

func (u *Unread) gather() {
	response, err := u.api.Request("messages.getConversations", map[string]string{
		"count":    "0",
		"group_id": u.groupId,
	})
	if err != nil {
		log.Println("Could not load unread messages", err)
		return
	}
	unread := json.Get(response, "response", "unread_count").ToInt()
	fields := map[string]interface{}{
		"unread": unread,
	}
	point, err := influx.NewPoint("unread", nil, fields, time.Now())
	if err != nil {
		log.Println("Could not create point: ", err.Error())
	} else {
		u.storage.PointsChannel <- point
	}
}
