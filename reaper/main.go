package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

const envLogLevel = "LOG_LEVEL"
const defaultLogLevel = logrus.InfoLevel

func main() {
	logLevel := getLogLevel()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logLevel)

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
		return defaultLogLevel
	}

	return level
}
