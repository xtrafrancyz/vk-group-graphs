package main

import (
	"log"
	"os"
	"time"

	influx "github.com/influxdata/influxdb1-client/v2"
)

type Storage struct {
	PointsChannel     chan *influx.Point
	batchPointsConfig influx.BatchPointsConfig
	client            influx.Client
}

func NewStorage(url, database, retentionPolicy string) *Storage {
	storage := &Storage{
		PointsChannel: make(chan *influx.Point),
		batchPointsConfig: influx.BatchPointsConfig{
			Precision:       "s",
			Database:        database,
			RetentionPolicy: retentionPolicy,
		},
	}
	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr: url,
	})
	if err != nil {
		log.Println("Error creating InfluxDB Client: ", err.Error())
		os.Exit(0)
	}

	storage.client = influxClient
	return storage
}

func (s *Storage) Run() {
	bp, _ := influx.NewBatchPoints(s.batchPointsConfig)
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case p := <-s.PointsChannel:
				bp.AddPoint(p)
			case <-ticker.C:
				if len(bp.Points()) > 0 {
					log.Println("Write to InfluxDB")

					go func(points influx.BatchPoints) {
						err := s.client.Write(points)
						if err != nil {
							log.Println("Could not save points: " + err.Error())
						}
					}(bp)

					bp, _ = influx.NewBatchPoints(s.batchPointsConfig)
				}
			}
		}
	}()
}
