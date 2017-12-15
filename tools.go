package spantool

import (
	"cloud.google.com/go/spanner"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/go-errors/errors"
	"os"
	"reflect"
	"time"
)

type Logger interface {
	init() error
	log(func() float64)
}

type LoggerDatadog struct {
	StatsdAddr      string
	Namespace       string
	Tags            []string
	PollingDuration time.Duration
	stats           *statsd.Client
}

func (logger LoggerDatadog) init() error {
	if logger.Namespace == "" {
		logger.Namespace = "spanner."
	}
	if logger.PollingDuration.Seconds() == 0 {
		logger.PollingDuration = 5 * time.Minute
	}

	hostname, _ := os.Hostname()
	logger.Tags = append(logger.Tags, "hostname:"+hostname)

	stats, err := statsd.New(logger.StatsdAddr)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	stats.Namespace = logger.Namespace
	stats.Tags = logger.Tags
	logger.stats = stats
	return nil
}

func (logger LoggerDatadog) log(sessionsCount func() float64) {
	go func() {
		for _ = range time.Tick(logger.PollingDuration) {
			logger.stats.Gauge("sessions", sessionsCount(), nil, 1) // TODO: what should the val of rate be?
		}
	}()
}

type LoggerStdout struct {
	PollingDuration time.Duration
}

func (logger LoggerStdout) init() error {
	return nil
}

func (logger LoggerStdout) log(sessionsCount func() float64) {
	go func() {
		for _ = range time.Tick(logger.PollingDuration) {
			fmt.Println("Sessions count is", sessionsCount())
		}
	}()
}

func LogSessionsCount(spannerClient *spanner.Client, logger Logger) error {
	if err := logger.init(); err != nil {
		return errors.Wrap(err, 0)
	}

	sessionsCount := func() float64 {
		count := reflect.ValueOf(*spannerClient).
			FieldByName("idleSessions").
			Elem().
			FieldByName("hc").
			Elem().
			FieldByName("queue").
			FieldByName("sessions")
		return float64(count.Len())
	}

	logger.log(sessionsCount)

	return nil
}
