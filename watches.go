package main

import (
	"time"
	"strings"
	"strconv"
	"fmt"

	"github.com/fzzy/radix/redis"
)

type Watch interface {
	Interval() time.Duration
	RedisCommand() string
	StatsdTarget() string
	ProcessReply(*redis.Reply) (int, error)
}

type CommandWatch struct {
	command string
	target string
	interval time.Duration
}

func NewCommandWatch(metric Metric) CommandWatch {
	return CommandWatch{metric.Command, metric.Name, time.Duration(metric.Interval) * time.Millisecond}
}

func (cw CommandWatch) Interval() time.Duration {
	return cw.interval
}

func (cw CommandWatch) RedisCommand() string {
	return cw.command
}

func (cw CommandWatch) StatsdTarget() string {
	return cw.target
}

func (cw CommandWatch) ProcessReply(reply *redis.Reply) (val int, err error) {
	return reply.Int()
}

type InternalWatch struct {
	key string
	target string
	interval time.Duration
}

func NewInternalWatch(metric IMetric) InternalWatch {
	return InternalWatch{metric.Key, metric.Name, time.Duration(metric.Interval) * time.Millisecond}
}

func (iw InternalWatch) Interval() time.Duration {
	return iw.interval
}

func (iw InternalWatch) RedisCommand() string {
	return "INFO"
}

func (iw InternalWatch) StatsdTarget() string {
	return iw.target
}

func (iw InternalWatch) ProcessReply(reply *redis.Reply) (val int, err error) {
	str, err := reply.Str()
	if err != nil {
		return
	}
	for _, byteLine := range strings.Split(str, "\r\n") {
		line := string(byteLine)
		byParts := strings.Split(line, ":")
		if len(byParts) != 2 {
			continue
		}

		if byParts[0] == iw.key {
			valf, err := strconv.ParseFloat(byParts[1], 64)
			if err != nil {
				return 0, err
			}
			val = int(valf)
			return val, err
		}
	}
	err = fmt.Errorf("No key ", iw.key, " found in INFO reply")
	return
}