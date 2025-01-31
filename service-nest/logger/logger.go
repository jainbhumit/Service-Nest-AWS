package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"time"
)

var log = logrus.New()

func InitLogger() {
	// Set the logger to write to stdout
	log.SetOutput(os.Stdout)

	logrus.SetFormatter(&logrus.JSONFormatter{})
	// Set logger to JSON format
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339, // Time format
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
	})

	// Set log level to Info by default
	log.SetLevel(logrus.InfoLevel)
}

// Info logs information with the given message
func Info(msg string, fields map[string]interface{}) {
	entry := log.WithFields(logrus.Fields{
		"caller": getCaller(),
	})
	for k, v := range fields {
		entry = entry.WithField(k, v)
	}
	entry.Info(msg)
}

// Error logs an error with the given message
func Error(msg string, fields map[string]interface{}) {
	entry := log.WithFields(logrus.Fields{
		"caller": getCaller(),
	})
	for k, v := range fields {
		entry = entry.WithField(k, v)
	}
	entry.Error(msg)
}

// Helper function to retrieve the caller
func getCaller() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	return fn.Name()
}
