package main

import (
	"github.com/gofiber/fiber/v2"
	logging "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"time"
)

func initRouter() *fiber.App{

	app := fiber.New()

	app.Use(requestid.New())

	app.Use(logging.New(logging.Config{
		Format:"${pid} ${locals:requestid} ${status} - ${method} ${path}\n",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		logger.Info(c.Locals("requestid"))
		return c.SendString("OK")
	})

	v1 := app.Group("/v1",setBackgroundContext)
	v1.Use(validateInputs)
	//yahoo finance group
	v1Yf := v1.Group("/yf")
	v1Yf.Post("/conversations/csv", timeout.New(yfConversationsHandler, time.Duration(getEnvInt64("REQUEST_TIMEOUT")) * time.Second))
	v1Yf.Post("/coversations/analyze", timeout.New(GenerateTextSentiments, time.Duration(getEnvInt64("REQUEST_TIMEOUT")) * time.Second))
	//twitter group
	v1Twitter := v1.Group("/twitter")
	v1Twitter.Post("/tweets/csv", timeout.New(twitterRecentTweetsHandler, time.Duration(getEnvInt64("REQUEST_TIMEOUT")) * time.Second))
	//reddit group
	v1Reddit := v1.Group("/reddit")
	v1RedditSearch := v1Reddit.Group("/:search_type",validateRedditInput)
	v1RedditSearch.Post("/csv", timeout.New(redditSearchHandler, time.Duration(getEnvInt64("REQUEST_TIMEOUT")) * time.Second))
	return app
}
