package main

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Storage struct {
	writer api.WriteAPI
	client influxdb2.Client
}

func NewStorage(config ConfigInfluxdb) *Storage {
	s := &Storage{
		client: influxdb2.NewClient(config.Url, config.Token),
	}
	s.writer = s.client.WriteAPI(config.Org, config.Bucket)
	return s
}

func (s *Storage) AddPoint(point *write.Point) {
	s.writer.WritePoint(point)
}
