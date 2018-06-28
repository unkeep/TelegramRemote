package main

import (
	"bytes"
	"log"
	"os/exec"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type controller struct {
	bot        *tgbotapi.BotAPI
	botUpdates tgbotapi.UpdatesChannel
	cfg        *config
}

func createController(bot *tgbotapi.BotAPI, c *config) (*controller, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updateChan, err := bot.GetUpdatesChan(u)

	if err != nil {
		return nil, err
	}

	return &controller{bot, updateChan, c}, nil
}

func (c *controller) start() {
	log.Println("Handling chat updates...")

	for {
		select {
		case update := <-c.botUpdates:
			c.handleUpdate(update)
		}
	}
}

func (c *controller) handleUpdate(update tgbotapi.Update) {
	var commandStr string
	if update.Message.Text[0] == '/' {
		v, ok := c.cfg.Commands[update.Message.Text]
		if !ok {
			c.reply(update.Message, "unsupported command")
			return
		}
		commandStr = v
	} else {
		commandStr = update.Message.Text
	}

	log.Print(commandStr)

	words := strings.Split(commandStr, " ")
	cmd := exec.Command(words[0], words[1:]...)
	// cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		c.reply(update.Message, err.Error())
	} else {
		c.reply(update.Message, out.String())
	}
}

func (c *controller) reply(toMsg *tgbotapi.Message, text string) {
	c.bot.Send(tgbotapi.NewMessage(toMsg.Chat.ID, text))
}
