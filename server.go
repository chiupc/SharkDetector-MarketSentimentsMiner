package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()
var th *TwitterApiHandler

func initLogger() {
	//initialize logger
	logLevelPtr := flag.String("loglevel", "info", "trace,debug,info,warn,error")
	flag.Parse()
	logger.SetFormatter(&logrus.JSONFormatter{})
	switch logLevel := *logLevelPtr; logLevel {
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.Info(fmt.Sprintf("Log level is set to %s",*logLevelPtr))
	//logger.SetReportCaller(true)
}

func main() {
	//load environment variables from .env file
	godotenv.Load(".env")
	initLogger()
	app := initRouter()

	//generate uuid for context id
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		logger.Error(err)
	}
	//uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	//uuid := utils.UUID()
	//ctx := context.WithValue(context.Background(), "ctx-id", uuid)
	//logger.WithFields(logrus.Fields{
	//	"ctx-id": ctx.Value("ctx-id"),
	//}).Info("Test mining yahoo finance conversations.")

	//create new twitter_handler with maximum connection to each host
	th = NewTwitterApiHandler()
	th.initNetClient()

	app.Listen(":3000")
}
