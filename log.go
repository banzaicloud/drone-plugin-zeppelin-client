package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

}

func SetLogFormat(logFormat string) {
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		if logFormat != "text" {
			log.Warnf("Unknown log format %s, use default text", logFormat)
		}
		log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	}
}

func SetLogLevel(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Unable to set log level, use default (info)")
	} else {
		log.SetLevel(level)
	}
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Error(args ...interface{}) {
	log.Info(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
