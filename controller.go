package main

import (
	"bytes"
	"fmt"
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
	if !c.isAuthorizedUser(update.Message.From.UserName) {
		c.reply(update.Message, "Sorry, you are not authorized")
		return
	}

	if update.Message.Text == "/help" {
		c.replyToHelp(update.Message)
		return
	}

	var commandStr string
	if update.Message.Text[0] == '/' {
		v, ok := c.cfg.Commands[update.Message.Text]
		if !ok {
			c.reply(update.Message, "Unsupported command")
			return
		}
		commandStr = v
	} else {
		commandStr = update.Message.Text
	}

	log.Printf("executing '%s'", commandStr)

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
	mConf := tgbotapi.NewMessage(toMsg.Chat.ID, text)
	mConf.ParseMode = tgbotapi.ModeHTML
	c.bot.Send(mConf)
}

func (c *controller) isAuthorizedUser(user string) bool {
	for _, u := range c.cfg.WhiteList {
		if u == user {
			return true
		}
	}

	return false
}

func (c *controller) replyToHelp(toMsg *tgbotapi.Message) {
	var lines []string
	for name, script := range c.cfg.Commands {
		lines = append(lines, fmt.Sprintf("%s - <i>%s</i>", name, script))
	}

	c.reply(toMsg, strings.Join(lines, "\n"))
}
