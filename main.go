package main

import (
	"flag"
	"log"
	"os"

	"github.com/xtrafrancyz/vk-group-graphs/vkapi"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Bind     string
	Vk       ConfigVk
	Influxdb ConfigInfluxdb
}

type ConfigVk struct {
	Confirmation string
	Secret       string
	Token        string
	GroupId      string
}

type ConfigInfluxdb struct {
	Url    string
	Org    string
	Bucket string
	Token  string
}

func main() {
	configFile := flag.String("config", "config.yaml", "config file path")
	flag.Parse()

	var config Config
	if b, err := os.ReadFile(*configFile); err != nil {
		log.Fatalln("Unable to read config file", err)
	} else if err = yaml.Unmarshal(b, &config); err != nil {
		log.Fatalln("Unable to parse config", err)
	}

	storage := NewStorage(config.Influxdb)

	api := vkapi.CreateWithToken(config.Vk.Token, "5.103")

	counter := NewCounter(storage)
	NewUnread(storage, api, config.Vk.GroupId)

	webServer := &WebServer{
		confirmationKey: config.Vk.Confirmation,
		secretKey:       config.Vk.Secret,
		onMessageOut:    counter.OnMessageOut,
	}
	webServer.Listen(config.Bind)
}
