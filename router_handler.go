package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strings"
	"path"
)

func yfConversationsHandler(c *fiber.Ctx) error {
	c.Accepts("application/json")
	logger.Info(string(c.Body()))
	q := new(YFRequestConversations)
	if err := c.BodyParser(q); err == nil{
		//c.Locals("ctx").(context.Context) => convert to context.Context
		res, err := collectMessage(c.Locals("ctx").(context.Context), q.Quote, q.StartTime, q.EndTime)
		if res == nil && err == nil{
			c.Status(fiber.StatusNoContent)
			return c.SendString("Success")
		}else if err != nil{
			c.Status(fiber.StatusBadRequest)
			return c.SendString("Bad Request")
		}else{
			//return c.JSON(res)
			logger.Debug(path.Join("tmp",res.Filename+".csv"))
			//c.Attachment(path.Join("tmp",res.Filename+".csv"))
			return c.Download(path.Join("tmp",res.Filename+".csv"))
		}
	}
	c.Status(fiber.StatusBadRequest)
	return c.SendString("please check inputs")
}

func twitterRecentTweetsHandler(c *fiber.Ctx) error{
	c.Accepts("application/json")
	q := new(TwitterRecentSearch)
	if err := c.BodyParser(q); err == nil {
		th.twitterRecentSearch(c.Locals("ctx").(context.Context),q.Quote, q.StartTime, q.EndTime)
		return c.SendString("success")
	}else{
		c.Status(fiber.StatusBadRequest)
		return c.SendString("please check query inputs")
	}
	return nil
}

func redditSearchHandler(c *fiber.Ctx) error{

	c.Accepts("application/json")
	q := new(RedditRecentSearch)
	if err := c.BodyParser(q); err == nil {
		err = r.redditSubmissionSearch(c.Locals("ctx").(context.Context),c.Params("search_type"),q.Quote, q.StartTime, q.EndTime)
		if err != nil {
			if err.Error() == "no result" {
				return c.SendString("no result found")
			}
		}
		return c.SendString("success")
	}else{
		return fiber.NewError(fiber.StatusBadRequest,"please check query inputs")
	}
	return nil
}

func validateRedditInput(c *fiber.Ctx) error{
	logger.Debug(c.Params("search_type"))
	logger.Debug(redditSearchTypes)
	logger.Debug(redditSearchTypes[c.Params("search_type")])
	if !redditSearchTypes[c.Params("search_type")]{
		logger.Debug("return false")
		return fiber.NewError(fiber.StatusNotFound,fmt.Sprintf("%s is not a valid reddit post type",c.Params("search_type")))
	}
	return c.Next()
}

//middleware
func setBackgroundContext(c *fiber.Ctx) error {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		logger.Error(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	//uuid := utils.UUID()
	logger.Debug(fmt.Sprintf("Setting context id... %s",uuid))
	ctx := context.WithValue(context.Background(), "ctx-id", uuid)
	c.Locals("ctx",ctx)
	return c.Next()
}


func validateInputs(c *fiber.Ctx) error{
	q := new(TimeRange)
	//default max time range to 1 week. can be overridden in .env file for each app
	maxTimeRange := int64(604800)
	if strings.Contains(c.Path(),"twitter"){
		maxTimeRange = getEnvInt64("TWITTER_MAXTIMERANGE")
	}else if strings.Contains(c.Path(),"yf"){
		maxTimeRange = getEnvInt64("YF_MAXTIMERANGE")
	}else if strings.Contains(c.Path(),"reddit"){
		maxTimeRange = getEnvInt64("REDDIT_MAXTIMERANGE")
	}
	if err := c.BodyParser(q); err == nil {
		errorResponses := validateTimeRange(q.StartTime,q.EndTime,maxTimeRange)
		if len(errorResponses) > 0{
			return fiber.NewError(fiber.StatusBadRequest,errorResponses[0].Value)
		}else{
			return c.Next()
		}
	}else{
		return fiber.NewError(fiber.StatusBadRequest,"Please check input parameters")
	}
}