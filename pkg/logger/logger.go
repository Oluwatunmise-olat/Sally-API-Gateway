package logger

import (
	"github.com/joho/godotenv"
	logrus_papertrail "github.com/polds/logrus-papertrail-hook"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var (
	Log *logrus.Logger
)

func init() {
	godotenv.Load()

	Log = logrus.New()
	Log.Out = os.Stdout
	Log.Formatter = &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     true,
	}
	Log.Level = logrus.InfoLevel
	Log.ReportCaller = false

	port, _ := strconv.Atoi(os.Getenv("PAPERTRAIL_PORT"))
	hook, _ := logrus_papertrail.NewPapertrailHook(&logrus_papertrail.Hook{
		Host:    os.Getenv("PAPERTRAIL_HOST"),
		Port:    port,
		Appname: "sally-api-gateway",
	})

	hook.SetLevels([]logrus.Level{logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel})

	Log.AddHook(hook)

}
