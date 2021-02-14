package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RedditApiHandler struct {
	Transport *http.Transport
	NetClient *http.Client
	TimeOut int
	MaxTimeRange int64
	MaxResult int
	Headers []string
	RateLimiter *rate.Limiter
	QueryTimeInterval int64
}

type RedditBuffer struct{
	Buffer []byte
	StatusCode int
}

func NewRedditApiHandler() *RedditApiHandler {
	return &RedditApiHandler{
		Transport: &http.Transport{
			Proxy:                  http.ProxyFromEnvironment,
			TLSHandshakeTimeout:    60 * time.Second,
			MaxIdleConns:           0,
			MaxIdleConnsPerHost:    100,
			MaxConnsPerHost:        100,
			IdleConnTimeout:        time.Duration(getEnvInt("REDDIT_TIMEOUT")) *  time.Second,
		},
		TimeOut: getEnvInt("REDDIT_TIMEOUT"),
		MaxTimeRange: getEnvInt64("REDDIT_MAXTIMERANGE"),
		MaxResult: getEnvInt("REDDIT_MAXRESULT"),
		RateLimiter: rate.NewLimiter(rate.Limit(getEnvInt("REDDIT_RATELIMIT_TASKPERSEC")), getEnvInt("REDDIT_RATELIMIT_MAXBURST")),
		QueryTimeInterval: getEnvInt64("REDDIT_QUERY_TIMEINTERVAL"),
	}
}

func (r *RedditApiHandler) initNetClient(){
	r.NetClient = &http.Client{
		Transport: r.Transport,
		Timeout:   time.Duration(getEnvInt("TWITTER_TIMEOUT")) *  time.Second,
	}
}

func (r *RedditApiHandler) redditSubmissionSearch(ctx context.Context, redditSearchType string, quote string, startTime int64, endTime int64) error{
	logger.Debug(redditTextFields)
	log := logger.WithFields(logrus.Fields{
		"ctx-id": ctx.Value("ctx-id"),
		"func": "redditSubmissionSearch",
	})
	//set headers based on search type
	r.Headers = redditSearchTypeFields[redditSearchType]
	baseUrl, err := url.Parse("https://api.pushshift.io/")
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return err
	}
	baseUrl.Path += fmt.Sprintf("reddit/search/%s",redditSearchType)
	query := fmt.Sprintf("%s",quote)
	params := url.Values{}
	params.Add("q", query)
	//params.Add("fields", os.Getenv("REDDIT_FIELDS"))
	params.Add("after",strconv.Itoa(int(startTime)))
	params.Add("before",strconv.Itoa(int(endTime)))
	params.Add("size","500")
	baseUrl.RawQuery = params.Encode()
	log.Debug("URL: ",baseUrl.String())
	fields := fmt.Sprintf("&fields=%s", strings.Join(r.Headers,","))
	req, err := http.NewRequest("GET", baseUrl.String() + fields ,nil)
	resp, err := r.NetClient.Do(req)
	if err != nil {
		log.Error("An Error Occured %v", err)
		return err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error(err)
		return err
	}
	filename_ := generateFileName(ctx, quote, int(startTime), int(endTime))

	//check if all result is returned from single query, possibly the returned message does not cover all results
	var jsonBody_ map[string]interface{}
	json.Unmarshal(body, &jsonBody_)
	isAllReturned := true
	if len(jsonBody_["data"].([]interface{})) == 0 { return errors.New("no result") }
	if len(jsonBody_["data"].([]interface{})) >= r.MaxResult { isAllReturned = false }
	//distribute the time range, if weeks then daily, if days the hourly, if hourly then 5-minute interval
	//timeRange := endTime - startTime
	if isAllReturned{
		log.Debug(fmt.Sprintf("Data length is %d, all result returned",len(jsonBody_["data"].([]interface{}))))
		r.jsonToCSV(ctx, filename_, body, true)
	}else{
		log.Debug(fmt.Sprintf("Data length is %d, not all result returned",len(jsonBody_["data"].([]interface{}))))
		inCludeHeader := true
		timeRange := endTime - startTime
		startTime_ := startTime
		endTime_ := startTime
		interval := r.QueryTimeInterval
		if timeRange <= 3600{
			interval = 300
		}else if timeRange > 3600 && timeRange <= 86400 {
			interval = 900
		}else if timeRange > 86400{
			interval = 3600
		}
		iter := 0
		var wg sync.WaitGroup
		var chans []chan RedditBuffer
		//var buf []byte
		for {
			//rate limit
			now := time.Now()
			rv := r.RateLimiter.Reserve()
			if !rv.OK() {
				logger.Error("rate limiter burst threshold is exceeded")
			}
			delay := rv.DelayFrom(now)
			logger.Debug(fmt.Sprintf("Rate limiter delay: %d",delay))
			params.Set("after", strconv.Itoa(int(startTime_)))
			endTime_ = endTime_ + int64(interval)
			if endTime_ > endTime{
				endTime_ = endTime
			}
			log.Debug(fmt.Sprintf("Iteration: %d, Start Time: %d, End Time: %d", iter, startTime_, endTime_))
			params.Set("before", strconv.Itoa(int(endTime_)))
			baseUrl.RawQuery = params.Encode()
			req, _ = http.NewRequest("GET", baseUrl.String() + fields ,nil)
			log.Debug("URL: ", req.URL.String())
			wg.Add(1)
			chans = append(chans, make(chan RedditBuffer))
			go r.redditNetWorker(ctx, iter, &wg, req, chans[iter])
			startTime_ = endTime_
			iter = iter + 1
			if endTime_ >= endTime {
				break
			}
			time.Sleep(delay)
		}
		for i, channel := range chans{
			redditBuffer := <-channel
			log.Info(fmt.Sprintf("Channel %d done, status code: %d, buf: %s", i, redditBuffer.StatusCode, string(redditBuffer.Buffer)))
			if redditBuffer.StatusCode == 200 {
				r.jsonToCSV(ctx, filename_, redditBuffer.Buffer, inCludeHeader)
			}
			inCludeHeader = false
		}
		wg.Wait()
		log.Debug("All workers completed.")
	}
	return nil
}

func (r *RedditApiHandler) redditNetWorker(ctx context.Context, id int, wg *sync.WaitGroup, req *http.Request, ch chan <- RedditBuffer){
	log := logger.WithFields(logrus.Fields{
		"ctx-id": ctx.Value("ctx-id"),
		"func": "redditNetWorker",
	})
	defer wg.Done()
	var buf []byte
	log.Debug(fmt.Sprintf("Worker %d is starting\n",id))
	resp, err := r.NetClient.Do(req)
	if err != nil {
		log.Error(err)
		ch <- RedditBuffer{
			Buffer:     nil,
			StatusCode: 500,
		}
	}
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		ch <- RedditBuffer{
			Buffer:     nil,
			StatusCode: 500,
		}
	}
	ch <- RedditBuffer{
		Buffer:     buf,
		StatusCode: resp.StatusCode,
	}
}

func (r *RedditApiHandler) jsonToCSV(ctx context.Context, filename string, body []byte, includeHeader bool)error{
	log := logger.WithFields(logrus.Fields{
		"ctx-id": ctx.Value("ctx-id"),
		"func": "jsonToCSV",
	})
	var lines []string
	var jsonBody map[string]interface{}
	filename = filename + ".csv"
	//log.Debug(string(body))
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		log.Error(err)
		return err
	}
	data := jsonBody["data"].([]interface{})

	//get the fields from the data to form header line
	if includeHeader { lines = append(lines, strings.Join(r.Headers, ",")) }
	for _, data_ := range data{
		//ensure the line is always in correct sequence
		var strs []string
		for _, header := range r.Headers{
			if redditTextFields[header]{ //if field is text
				strs = append(strs, fmt.Sprintf("\"%s\"",r.cleanup(data_.(map[string]interface{})[header].(string))))
			}else {
				strs = append(strs, fmt.Sprintf("%v", data_.(map[string]interface{})[header]))
			}
		}
		log.Debug(strs)
		line := strings.Join(strs, ",")
		lines = append(lines,line)
	}
	log.Debug(lines)
	writeToCSV(ctx, filename, lines)
	return nil
}

func (r *RedditApiHandler) cleanup(s string) string{
	//remove '\n'
	s = strings.Replace(s, "\n", " ", -1)
	s = strings.Replace(s, "\"", "'", -1)
	return s
}