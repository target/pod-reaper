package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	reaper := newReaper()
	reaper.harvest()
	logrus.Info("pod reaper is exiting")
}
