package main

import (
	"context"
	"crypto/rand"
	"fmt"
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
	v1.Post("/yf/conversations/csv", timeout.New(yfConversations, 60 * time.Second))
	v1.Post("/twitter/tweets/csv", timeout.New(twitterRecentTweets, 60 * time.Second))

	return app
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