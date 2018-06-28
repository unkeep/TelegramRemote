package main

import (
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	container.Provide(loadConfig)
	container.Provide(createBotAPI)
	container.Provide(createController)

	err := container.Invoke(func(c *controller) {
		c.start()
	})

	if err != nil {
		log.Panic(err)
	}
}

func createBotAPI(c *config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(c.BotToken)
	if err == nil {
		log.Printf("Authorized on bot account %s", bot.Self.UserName)
	}

	return bot, err
}
