package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var (
	Log = &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			PrettyPrint:     false,
		},
		Level:        logrus.InfoLevel,
		ReportCaller: false,
	}
)
