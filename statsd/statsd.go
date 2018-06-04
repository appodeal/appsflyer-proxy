package statsd

import (
	"fmt"
	"time"
	"net"
	"io"
	"strings"
)

type StatsD struct {
	address string
	project string
	queue   chan string
}

func NewStatsD(address string, project string, buffer int) (*StatsD, error) {

	parts := strings.SplitN(address,"/",2)

	sd := StatsD{
		address: parts[0],
		project: parts[1],
		queue: make(chan string, buffer),
	}
	go sd.sender()
	
	return &sd, nil
}

func (sd *StatsD) sender() {
	for s := range sd.queue {
		if conn, err := net.Dial("udp", sd.address); err == nil {
			io.WriteString(conn, s)
			conn.Close()
		}
	}
}

func (sd *StatsD) Close() {
	close(sd.queue)
}

func (sd *StatsD) Count(metric string, value int) {
	sd.queue <- fmt.Sprintf("%s.%s:%d|c", sd.project, metric, value)
}

func (sd *StatsD) Time(metric string, took time.Duration) {
	sd.queue <- fmt.Sprintf("%s.%s:%d|ms", sd.project, metric, took/1e6)
}

func (sd *StatsD) Gauge(metric string, value int) {
	sd.queue <- fmt.Sprintf("%s.%s:%d|g", sd.project, metric, value)
}
