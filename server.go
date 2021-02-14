package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

var logger = logrus.New()
var th *TwitterApiHandler
var r *RedditApiHandler

var redditSearchTypes map[string]bool
var redditTextFields map[string]bool
var redditSearchTypeFields map[string][]string

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

func setConstants(){
	//reddit
	redditSearchTypes = make(map[string]bool)
	envRedditSearchTypes := os.Getenv("REDDIT_SEARCH_TYPES")
	envRedditSearchTypes_ := strings.Split(envRedditSearchTypes,",")
	redditSearchTypeFields = make(map[string][]string)
	for _, redditSearchType := range envRedditSearchTypes_{
		redditSearchTypes[redditSearchType] = true
		redditSearchTypeFields[redditSearchType] = strings.Split(os.Getenv("REDDIT_" + strings.ToUpper(redditSearchType) + "_FIELDS"),",")
	}

	redditTextFields = make(map[string]bool)
	envRedditTextFields := os.Getenv("REDDIT_TEXT_FIELDS")
	envRedditTextFields_ := strings.Split(envRedditTextFields,",")
	for _, redditTextField := range envRedditTextFields_{
		redditTextFields[redditTextField] = true
	}

}

func main() {
	//load environment variables from .env file
	godotenv.Load(".env")
	setConstants()
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
	r = NewRedditApiHandler()
	r.initNetClient()

	app.Listen(":3000")
}
