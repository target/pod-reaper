package main

import (
	"os"
	"reflect"
	"testing"

	joonix "github.com/joonix/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var levelTests = []struct {
	str   string
	level logrus.Level
}{
	{"PANIC", logrus.PanicLevel},
	{"FATAL", logrus.FatalLevel},
	{"ERROR", logrus.ErrorLevel},
	{"WARNING", logrus.WarnLevel},
	{"DEBUG", logrus.DebugLevel},
	{"INFO", logrus.InfoLevel},
}

func TestGetLogLevel(t *testing.T) {
	t.Run("default not set", func(t *testing.T) {
		os.Clearenv()
		level := getLogLevel()
		assert.Equal(t, level, defaultLogLevel)
	})
	t.Run("default invalid", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envLogLevel, "foo")
		level := getLogLevel()
		assert.Equal(t, level, defaultLogLevel)
	})
	for _, tt := range levelTests {
		t.Run(tt.str, func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envLogLevel, tt.str)
			level := getLogLevel()
			assert.Equal(t, level, tt.level)
		})
	}
}

func TestGetLogFormat(t *testing.T) {
	t.Run("default not set", func(t *testing.T) {
		os.Clearenv()
		format := getLogFormat()
		assert.Equal(t, reflect.TypeOf(format), reflect.TypeOf(&logrus.JSONFormatter{}))
	})
	t.Run("default invalid", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envLogFormat, "foo")
		format := getLogFormat()
		assert.Equal(t, reflect.TypeOf(format), reflect.TypeOf(&logrus.JSONFormatter{}))
	})
	t.Run("logrus", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envLogFormat, "Logrus")
		format := getLogFormat()
		assert.Equal(t, reflect.TypeOf(format), reflect.TypeOf(&logrus.JSONFormatter{}))
	})
	t.Run("fluentd", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envLogFormat, "Fluentd")
		format := getLogFormat()
		assert.Equal(t, reflect.TypeOf(format), reflect.TypeOf(&joonix.FluentdFormatter{}))
	})
}
