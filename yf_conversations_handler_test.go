package main

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
)

func TestGetMessageBoardId(t *testing.T) {
	messageBoardId, err := getMessageBoardId("AAPL")
	if err == nil{
		if messageBoardId != "finmb_24937"{
			t.Errorf("Message board Id was incorrect, got: %s, want: %s.", messageBoardId, "finmb_24937")
		}
	}else{
		t.Errorf(err.Error())
	}
}

func TestGetMessageBoardId_InvalidQuote(t *testing.T) {
	messageBoardId, err := getMessageBoardId("AAPLHAHA")
	if err != nil && messageBoardId == ""{
		fmt.Printf("Expected error %s\n", err.Error())
	}else{
		t.Errorf("Message board Id was incorrect, got: %s, want: %s.", messageBoardId, "")
	}
}

func TestCollectMessage( *testing.T){
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		logger.Error(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	ctx := context.WithValue(context.Background(), "ctx-id", uuid)
	res, err := collectMessage(ctx, "AAPL", 1608008280, 1609131480)
	res, err = collectMessage(ctx, "MU", 1608008280, 1609131480)
	if res != nil && err != nil{
		logger.Error(err)
	}
}

func TestCollectMessage_InvalidStartEndTime( *testing.T){
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		logger.Error(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	ctx := context.WithValue(context.Background(), "ctx-id", uuid)
	res, err := collectMessage(ctx, "AAPL", 1609131480, 1608008280)
	if res != nil && err == nil{
		logger.Info("Error of invalid inputs is expected")
	}
}
