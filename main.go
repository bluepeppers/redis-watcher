package main

import (
	"log"
	"flag"
	"strings"
	"time"
	"os"
	"io"

	"github.com/quipo/statsd"
	"github.com/fzzy/radix/extra/pool"
)

var redisAddr = flag.String("redis", "127.0.0.1:6379", "The address that the redis to be monitored can be found on")
var statsdAddr = flag.String("statsd", "127.0.0.1:8125", "The address that the statsd to be published to can be found")
var configLocation = flag.String("config", "", "The location of the config file")

func main() {
	flag.Parse()

	var confFile io.ReadCloser
	var err error
	if *configLocation == "" {
		confFile, err = FindConfig()
	} else {
		confFile, err = os.Open(*configLocation)
	}

	if err != nil {
		// TODO
		log.Fatal("Failed to load config file: ", err)
	}

	conf, err := LoadConfig(confFile)
	if err != nil {
		log.Fatal("Failed to decode config file: ", err)
	}

	redisPool := pool.NewOrEmptyPool("tcp", *redisAddr, 12)

	watches := make([]Watch, 0, len(conf.Metrics) + len(conf.Internal))
	for _, metric := range conf.Metrics {
		watch := NewCommandWatch(metric)
		watches = append(watches, watch)
	}
	for _, metric := range conf.Internal {
		watch := NewInternalWatch(metric)
		watches = append(watches, watch)
	}

	MonitorWatches(watches, redisPool)
}

func MonitorWatches(watches []Watch, redisPool *pool.Pool) {
	dueChan := make(chan Watch, len(watches))
	for _, watch := range watches {
		go func(watch Watch) {
			for {
				time.Sleep(watch.Interval())
				dueChan <- watch
			}
		} (watch)
	}

	for {
		toRun := <- dueChan
		go ExecuteWatch(toRun, redisPool)
	}
}

func ExecuteWatch(watch Watch, redisPool *pool.Pool) {
	statsdClient := statsd.NewStatsdClient(*statsdAddr, "udp")
	err := statsdClient.CreateSocket()
	if err != nil {
		log.Println("Failed to create statsd socket:", err)
		return
	}
	defer statsdClient.Close()

	conn, err := redisPool.Get()
	if err != nil {
		log.Println("Failed to redis conn: ", err)
		return
	}

	// TODO
	rArgs := strings.Split(watch.RedisCommand(), " ")
	args := make([]interface{}, len(rArgs) - 1)
	for i, v := range rArgs[1:] {
		args[i] = v
	}

	reply := conn.Cmd(rArgs[0], args...)
	if reply.Err != nil {
		log.Println("Failed to run ", watch.RedisCommand(), ": ", reply.Err)
		conn.Close()
		return
	}
	redisPool.Put(conn)
	
	val, err := watch.ProcessReply(reply)
	if err != nil {
		log.Println("Failed to process reply: ", err)
		return
	}

	err = statsdClient.Gauge(watch.StatsdTarget(), int64(val))
	if err != nil {
		log.Println("Failed to store value:", err)
		return
	}

	log.Println(watch.StatsdTarget(), "->", int64(val))
}