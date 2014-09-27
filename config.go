package main

import (
	"os"
	"io"
	"encoding/json"
	"fmt"
)

type Config struct {
	ReportInterval int `json:"report-interval"`
	Metrics []Metric `json:"metrics"`
	Internal []IMetric `json:"internal"`
}

type Metric struct {
	Name string `json:"name"`
	Command string `json:"command"`
	Interval int `json:"report-interval"`
}

type IMetric struct {
	Name string `json:"name"`
	Key string `json:"key"`
	Interval int `json:"report-interval"`
}

var CONFIG_LOCATIONS = []string{
	"redis-watcher.json",
	"/etc/redis-watcher.json",
}

func FindConfig() (file io.ReadCloser, err error) {
	for _, filename := range CONFIG_LOCATIONS {
		file, err = os.Open(filename)
		if err != nil && os.IsNotExist(err) {
			continue
		}
		
		return
	}
	return nil, fmt.Errorf("Could not find config file")
}

func LoadConfig(file io.ReadCloser) (conf Config, err error) {
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)

	if err != nil {
		return
	}
	
	if conf.ReportInterval == 0 {
		conf.ReportInterval = 1000
	}
	for i := range conf.Metrics {
		if conf.Metrics[i].Interval == 0 {
			conf.Metrics[i].Interval = conf.ReportInterval
		}
	}
	for i := range conf.Internal {
		if conf.Internal[i].Interval == 0 {
			conf.Internal[i].Interval = conf.ReportInterval
		}
	}
	return
}
