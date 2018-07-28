package main

import (
	"log"
	"strconv"

	"github.com/xtrafrancyz/vk-group-graphs/vkapi"
)

type SpamFilter struct {
	api *vkapi.Api
}

func (sf *SpamFilter) OnWallReply(object map[string]interface{}) {
	if object["text"].(string) == "" && object["reply_to_user"] == nil {
		if attachmentObj, ok := object["attachments"]; ok {
			attachments := attachmentObj.([]interface{})
			if len(attachments) == 1 && attachments[0].(map[string]interface{})["type"] == "video" {
				_, err := sf.api.Request("wall.deleteComment", map[string]string{
					"owner_id":   strconv.Itoa(int(object["post_owner_id"].(float64))),
					"comment_id": strconv.Itoa(int(object["id"].(float64))),
				})
				if err != nil {
					log.Println(err)
				} else {
					log.Println("Deleted comment with video")
				}
			}
		}
	}
}
