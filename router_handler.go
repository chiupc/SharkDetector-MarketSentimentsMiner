package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

func yfConversations(c *fiber.Ctx) error {
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
			c.Attachment(res.Filename+".csv")
			return c.JSON(res)
		}
	}
	c.Status(fiber.StatusBadRequest)
	return c.SendString("please check inputs")
}

func twitterRecentTweets(c *fiber.Ctx) error{
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