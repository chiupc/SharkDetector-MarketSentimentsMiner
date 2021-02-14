package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//Attempts to authenticate to twitter developer api, if fails then don't accept request
type TwitterApiHandler struct {
	Transport *http.Transport
	NetClient *http.Client
	TimeOut int
	MaxTimeRange int64
}

func NewTwitterApiHandler() *TwitterApiHandler {
	return &TwitterApiHandler{
		Transport: &http.Transport{
			Proxy:                  http.ProxyFromEnvironment,
			TLSHandshakeTimeout:    60 * time.Second,
			MaxIdleConns:           0,
			MaxIdleConnsPerHost:    100,
			MaxConnsPerHost:        100,
			IdleConnTimeout:        time.Duration(getEnvInt("TWITTER_TIMEOUT")) *  time.Second,
		},
		TimeOut: func() int {
			timeout, _ := strconv.Atoi(os.Getenv("TWITTER_TIMEOUT"))
			return timeout
		}(),
		MaxTimeRange: func() int64 {
			timeout, _ := strconv.Atoi(os.Getenv("TWITTER_MAXTIMERANGE"))
			return int64(timeout)
		}(),
	}
}

func (t *TwitterApiHandler) initNetClient(){
	t.NetClient = &http.Client{
		Transport:     t.Transport,
		Timeout:       time.Duration(getEnvInt("TWITTER_TIMEOUT")) *  time.Second,
	}
}

func (t *TwitterApiHandler) twitterRecentSearch(ctx context.Context, quote string, startTime int64, endTime int64) error{
	log := logger.WithFields(logrus.Fields{
		"ctx-id": ctx.Value("ctx-id"),
		"func": "twitterRecentSearch",
	})
	baseUrl, err := url.Parse("https://api.twitter.com/2/")
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return err
	}
	baseUrl.Path += "tweets/search/recent"
	query := fmt.Sprintf("#%s -is:retweet lang:en",quote)
	params := url.Values{}
	params.Add("query", query)
	params.Add("tweet.fields", os.Getenv("TWITTER_FIELDS"))
	params.Add("start_time",time.Unix(startTime,0).UTC().Format(time.RFC3339))
	params.Add("end_time",time.Unix(endTime,0).UTC().Format(time.RFC3339))
	params.Add("max_results","100")
	baseUrl.RawQuery = params.Encode()
	log.Debug("URL: ",baseUrl.String())
	req, err := http.NewRequest("GET", baseUrl.String() ,nil)
	req.Header.Add("Authorization",fmt.Sprintf("Bearer %s",os.Getenv("TWITTER_BEARER")))
	resp, err := t.NetClient.Do(req)
	if err != nil {
		logger.Error("An Error Occured %v", err)
		return err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}
	//check if next_token exists, if yes, continue to query
	filename_ := generateFileName(ctx, quote, int(startTime), int(endTime))
	t.jsonToCSV(ctx, filename_, body, true)
	nextToken, exists := t.getNextToken(body)
	if exists {
		params.Add("next_token", nextToken)
		for {
			logger.Debug(fmt.Sprintf("next_token found '%s'", nextToken))
			params.Set("next_token", nextToken)
			baseUrl.RawQuery = params.Encode()
			log.Debug("URL: ", baseUrl.String())
			req.URL = baseUrl
			resp, err = t.NetClient.Do(req)
			if err != nil {
				logger.Error(err)
				return err
			}
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error(err)
				return err
			}
			t.jsonToCSV(ctx, filename_, body, false)
			nextToken, exists = t.getNextToken(body)
			if !exists {
				break
			}
		}
	}
	return nil
}

func (t *TwitterApiHandler) getNextToken(body []byte)(string,bool){
	var jsonBody map[string]interface{}
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		logger.Error(err)
		return "", false
	}
	metadata := jsonBody["meta"].(map[string]interface{})
	nextToken, exists := metadata["next_token"].(string)
	return nextToken, exists
}

func (t *TwitterApiHandler) jsonToCSV(ctx context.Context, filename string, body []byte, includeHeader bool)error{
	var headers []string
	var lines []string
	var jsonBody map[string]interface{}
	filename = filename + ".csv"
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug(jsonBody)
	data := jsonBody["data"].([]interface{})
	logger.Debug(data)
	for field := range data[0].(map[string]interface{}) {
		headers = append(headers, field)
	}
	sort.Strings(headers)
	logger.Debug(headers)
	//get the fields from the data to form header line
	if includeHeader { lines = append(lines, strings.Join(headers, ",")) }
	for _, data_ := range data{
		//ensure the line is always in correct sequence
		var strs []string
		for _, header := range headers{
			if header == "text"{
				strs = append(strs, t.cleanup(fmt.Sprintf("\"%v\"",data_.(map[string]interface{})[header])))
			}else {
				strs = append(strs, t.cleanup(fmt.Sprintf("%v", data_.(map[string]interface{})[header])))
			}
		}
		logger.Debug(strs)
		line := strings.Join(strs, ",")
		lines = append(lines,line)
	}
	logger.Debug(lines)
	writeToCSV(ctx, filename, lines)
	return nil
}

func (t *TwitterApiHandler) cleanup(s string) string{
	//remove '\n'
	s = strings.Replace(s, "\n", " ", -1)
	return s
}
