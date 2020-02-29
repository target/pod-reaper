package main

import (
	"os"

	joonix "github.com/joonix/log"
	"github.com/sirupsen/logrus"
)

const envLogLevel = "LOG_LEVEL"
const envLogFormat = "LOG_FORMAT"
const fluentdFormat = "Fluentd"
const logrusFormat = "Logrus"
const defaultLogLevel = logrus.InfoLevel

func main() {
	logLevel := getLogLevel()
	logrus.SetLevel(logLevel)
	logFormat := getLogFormat()
	logrus.SetFormatter(logFormat)

	reaper := newReaper()
	reaper.harvest()
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
