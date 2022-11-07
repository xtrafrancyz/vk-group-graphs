package main

import (
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/xtrafrancyz/vk-group-graphs/vkapi"
)

type respConversations struct {
	Response struct {
		UnreadCount int `json:"unread_count"`
	} `json:"response"`
}

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
	var response respConversations
	err := u.api.RequestJsonStruct("messages.getConversations", map[string]string{
		"count":    "0",
		"group_id": u.groupId,
	}, &response)
	if err != nil {
		log.Println("Could not load unread messages", err)
		return
	}
	u.storage.AddPoint(influxdb2.NewPoint(
		"unread",
		nil,
		map[string]any{
			"unread": response.Response.UnreadCount,
		},
		time.Now(),
	))
}
