package main

import (
	"os"

	joonix "github.com/joonix/log"
	"github.com/sirupsen/logrus"
	"github.com/target/pod-reaper/cmd/pod-reaper/app"
	"github.com/target/pod-reaper/internal/pkg/client"
)

const envLogLevel = "LOG_LEVEL"
const envLogFormat = "LOG_FORMAT"
const fluentdFormat = "Fluentd"
const logrusFormat = "Logrus"
const defaultLogLevel = logrus.InfoLevel

func main() {
	logFormat := getLogFormat()
	logrus.SetFormatter(logFormat)
	logLevel := getLogLevel()
	logrus.SetLevel(logLevel)

	clientset, err := client.CreateClient("")
	if err != nil {
		logrus.WithError(err).Panic("cannot create client")
		panic(err)
	}
	reaper := app.NewReaper(clientset)
	reaper.Harvest()
	logrus.Info("pod reaper is exiting")
}

func getLogLevel() logrus.Level {
	levelString, exists := os.LookupEnv(envLogLevel)
	if !exists {
		return defaultLogLevel
	}

	level, err := logrus.ParseLevel(levelString)
	if err != nil {
		logrus.Errorf("error parsing %s: %v", envLogLevel, err)
		return defaultLogLevel
	}

	return level
}

func getLogFormat() logrus.Formatter {
	formatString, exists := os.LookupEnv(envLogFormat)
	if !exists || formatString == logrusFormat {
		return &logrus.JSONFormatter{}
	} else if formatString == fluentdFormat {
		return &joonix.FluentdFormatter{}
	} else {
		logrus.Errorf("unknown %s: %v", envLogFormat, formatString)
		return &logrus.JSONFormatter{}
	}
}
