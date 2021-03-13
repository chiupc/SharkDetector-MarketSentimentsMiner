package main

import (
	"github.com/chiupc/sentiment_analytic/client_handler"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateTextSentiments(t *testing.T){
	client := client_handler.NewSentimentAnalyticGrpcClient()
	client_handler.GetTextSentiments(client, filepath.Join(os.Getenv("DATA_PATH"),"NIO_1614600519_1614686919_02af901ae00e.csv"),"userText")
}