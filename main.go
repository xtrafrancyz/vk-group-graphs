package main

import (
	"flag"
)

func main() {
	var webHost = flag.String("webhost", "0.0.0.0:7424", "address to bind webserver")
	var confirmationKey = flag.String("confirmation-key", "123qwe", "confirmation key from vk")
	var secretKey = flag.String("secret-key", "123qwe", "secret key from vk")
	var influxUrl = flag.String("influx-url", "http://127.0.0.1:8086", "address of InfluxDB")
	var influxDatabase = flag.String("influx-database", "vk_graphs", "database name")
	var influxRetentionPolicy = flag.String("influx-rp", "a_year", "retention policy")

	flag.Parse()

	storage := NewStorage(*influxUrl, *influxDatabase, *influxRetentionPolicy)
	storage.Run()

	counter := NewCounter(storage)
	counter.ScheduleSave()
	NewWebServer(*webHost, *confirmationKey, *secretKey, counter.OnMessageSent)
}
