package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

const MaxBufferSize = 8192
const MaxMessageNum = 5000

func collectMessage(ctx context.Context, quote string, startTime int64, endTime int64) (*ConversationsResult, error){
	messageBoardId, err := getMessageBoardId(quote)
	if err != nil{
		return nil, err
	}
	if startTime >= endTime{
		return nil, errors.New("invalid input")
	}
	log := logger.WithFields(logrus.Fields{
		"ctx-id": ctx.Value("ctx-id"),
		"func": "collectMessage",
	})
	//initialize line array
	//lines := make([]string,MaxMessageNum)
	var lines []string
	headers := []string{"nickname","createdAt","region","userText"}
	headersJoined := strings.Join(headers,",")
	lines = append(lines, headersJoined)
	//build url
	host := "sg.finance.yahoo.com"
	q := "query=namespace%20%3D%20%22yahoo_finance%22%20and%20(contextId%3D%22"+ messageBoardId + "%22%20or%20tag%3D%22" + quote + "%22)"
	log.WithFields(logrus.Fields{"start_time":startTime,"end_time":endTime,"quote":quote}).Info("conversations mining start")
	offset := 0
	endTime_ := endTime
	for endTime_ >= startTime && len(lines) <= MaxMessageNum {
		log.Debug(fmt.Sprintf("mining messages at time (%d)",endTime_))
		parms := []string{"apiVersion=v1", "count=30", fmt.Sprintf("index=v=1:s=time:t=%d:off=%d", endTime, offset), "lang=en-SG",
			"namespace=yahoo_finance", "oauthConsumerKey=finance.oauth.client.canvass.prod.consumerKey",
			"oauthConsumerSecret=finance.oauth.client.canvass.prod.consumerSecret", q,
			"sortBy=createdAt", "type=null", "userActivity=true", "region=SG"}
		urlParms := strings.Join(parms, ";")
		path := "/_finance_doubledown/api/resource/canvass.getMessageList" + ";" + urlParms
		urlPath := "https://" + host + path
		resp, err := http.Get(urlPath)
		if err == nil && resp.StatusCode == 200 {
			var body []byte
			var jsonBody map[string]interface{}
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				json.Unmarshal(body, &jsonBody)
				log.WithField("response-status-code", resp.StatusCode).Debug(resp.Status)
				log.WithField("message-type", "response-body-json").Debug(jsonBody)
				//we need to get last index timestamp because yahoo finance conversations only allow max 30 messages
				messages := jsonBody["canvassMessages"].([]interface{})
				for _, message := range messages {
					//refer to https://gobyexample.com/json for unmarshalling json and type conversion
					message_ := message.(map[string]interface{})
					meta := message_["meta"].(map[string]interface{})
					//final extraction
					createdAt := int64(meta["createdAt"].(float64))
					log.Debug(fmt.Sprintf("createdAt: %d", createdAt))
					endTime_ = createdAt
					if createdAt >= startTime {
						createdAt := fmt.Sprintf("%d", createdAt)
						userText := message_["details"].(map[string]interface{})["userText"].(string)
						userText = strings.Replace(userText, "\n", " ", -1)
						nickname := meta["author"].(map[string]interface{})["nickname"].(string)
						region := meta["locale"].(map[string]interface{})["region"].(string)
						//build csv line
						strs := []string{nickname, createdAt, region, userText}
						line := strings.Join(strs, ",")
						lines = append(lines, line)
						log.Debug(line)
					}else{
						break
					}
				}
				offset = offset + 30
			} else {
				log.Error(err.Error())
			}
		} else {
			log.Error(fmt.Sprintf("Error Code: %d Message: %s", resp.StatusCode, resp.Body))
		}
		log.Debug(fmt.Sprintf("current number of lines: %d",len(lines)))
	}
	//test write csv file
	//TODO: add timestamp and unique id to file name
	if len(lines) > 1 {
		log.Debug(fmt.Sprintf("number of lines: %d", len(lines)))
		//generateFileName = {quote}_{startTime}_{endTime}_{last_substrings_of_uuid}
		filename := generateFileName(ctx, quote, int(startTime), int(endTime))
		res := ConversationsResult{
			Filename: filename,
			LineNum: len(lines),
		}
		writeToCSV(ctx, filename+".csv", lines)
		return &res, nil
	}else{
		log.Info("no result found")
	}
	return nil, nil
}

func getMessageBoardId(quote string) (string,error){
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s",quote)
	fmt.Println(url)
	resp, err := http.Get(url)
	if resp.StatusCode == 200 && err == nil{
		var body []byte
		var jsonBody map[string]interface{}
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			json.Unmarshal(body, &jsonBody)
		}
		result := jsonBody["quoteResponse"].(map[string]interface {})["result"].([]interface{})
		if len(result) > 0 {
			messageBoardId := result[0].(map[string]interface{})["messageBoardId"].(string)
			return messageBoardId, nil
		}
	}
	return "", errors.New("invalid quote")
}

