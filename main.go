package main

import (
	"flag"
)

func main() {
	var webHost = flag.String("webhost", "0.0.0.0:7424", "address to bind webserver")
	var confirmationKey = flag.String("confirmation-key", "0.0.0.0:7424", "confirmation key from vk")
	var secretKey = flag.String("secret-key", "0.0.0.0:7424", "secret key from vk")
	var influxUrl = flag.String("influx-url", "http://127.0.0.1:8086", "address of InfluxDB")
	var influxDatabase = flag.String("influx-database", "http://127.0.0.1:8086", "database name")
	var influxRetentionPolicy = flag.String("influx-rp", "a_day", "retention policy")

	flag.Parse()

	storage := NewStorage(*influxUrl, *influxDatabase, *influxRetentionPolicy)
	storage.Run()

	counter := NewCounter(storage)
	counter.ScheduleSave()
	NewWebServer(*webHost, *confirmationKey, *secretKey, counter.OnMessageSent)
}
