package main

import (
	"flag"

	"github.com/xtrafrancyz/vk-group-graphs/vkapi"
)

func main() {
	var webHost = flag.String("webhost", "0.0.0.0:7424", "address to bind webserver")
	var vkConfirmationKey = flag.String("vk-confirmation-key", "123qwe", "confirmation key from vk")
	var vkSecretKey = flag.String("vk-secret-key", "123qwe", "secret key from vk")
	var vkAccessToken = flag.String("vk-access-token", "123qwe", "access token from vk")
	var vkGroupId = flag.String("vk-group-id", "123", "id of vk group to get unread messages")
	var influxUrl = flag.String("influx-url", "http://127.0.0.1:8086", "address of InfluxDB")
	var influxDatabase = flag.String("influx-database", "vk_graphs", "database name")
	var influxRetentionPolicy = flag.String("influx-rp", "a_year", "retention policy")

	flag.Parse()

	storage := NewStorage(*influxUrl, *influxDatabase, *influxRetentionPolicy)
	storage.Run()

	api := vkapi.CreateWithToken(*vkAccessToken, "5.80")

	counter := NewCounter(storage)
	NewUnread(storage, api, *vkGroupId)

	spamFilter := SpamFilter{
		api: api,
	}

	webServer := &WebServer{
		confirmationKey: *vkConfirmationKey,
		secretKey:       *vkSecretKey,
		onMessageReply:  counter.OnMessageReply,
		onWallReply:     spamFilter.OnWallReply,
	}
	webServer.Listen(*webHost)
}
