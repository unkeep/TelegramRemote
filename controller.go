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
	for update := range c.botUpdates {
		c.handleUpdate(update)
	}
}

func (c *controller) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if !c.isAuthorizedUser(update.Message.From.UserName) {
		c.replyWithText(update.Message, "Sorry, you are not authorized")
		return
	}

	if update.Message.Text == "/help" {
		c.replyToHelp(update.Message)
		return
	}

	if strings.HasPrefix(update.Message.Text, "/file ") {
		path := strings.TrimPrefix(update.Message.Text, "/file ")
		c.replyWithFile(update.Message, path)
		return
	}

	var commandStr string
	if update.Message.Text[0] == '/' {
		v, ok := c.cfg.Commands[update.Message.Text]
		if !ok {
			c.replyWithText(update.Message, "Unsupported command")
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
		c.replyWithText(update.Message, err.Error())
	} else {
		c.replyWithText(update.Message, out.String())
	}
}

func (c *controller) replyWithText(toMsg *tgbotapi.Message, text string) {
	mConf := tgbotapi.NewMessage(toMsg.Chat.ID, text)
	mConf.ParseMode = tgbotapi.ModeHTML
	c.bot.Send(mConf)
}

func (c *controller) replyWithFile(toMsg *tgbotapi.Message, path string) {
	mConf := tgbotapi.NewDocumentUpload(toMsg.Chat.ID, path)
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

	c.replyWithText(toMsg, strings.Join(lines, "\n"))
}
