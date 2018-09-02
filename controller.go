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
	c.sendTypingAction(update.Message.Chat.ID)

	words := strings.Split(commandStr, " ")
	cmd := exec.Command(words[0], words[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		c.replyWithText(update.Message, err.Error())
	} else {
		log.Println("Out:", out.String())
		c.replyWithText(update.Message, out.String())
	}
}

func (c *controller) sendTypingAction(chatID int64) {
	mConf := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := c.bot.Send(mConf); err != nil {
		log.Println("sendTypingAction err:", err.Error())
	}
}

func (c *controller) replyWithText(toMsg *tgbotapi.Message, text string) {
	mConf := tgbotapi.NewMessage(toMsg.Chat.ID, text)
	if _, err := c.bot.Send(mConf); err != nil {
		log.Println("replyWithText err:", err.Error())
	}
}

func (c *controller) replyWithHTML(toMsg *tgbotapi.Message, html string) {
	mConf := tgbotapi.NewMessage(toMsg.Chat.ID, html)
	mConf.ParseMode = tgbotapi.ModeHTML
	if _, err := c.bot.Send(mConf); err != nil {
		log.Println("replyWithHTML err:", err.Error())
	}
}

func (c *controller) replyWithFile(toMsg *tgbotapi.Message, path string) {
	mConf := tgbotapi.NewDocumentUpload(toMsg.Chat.ID, path)
	c.bot.Send(mConf)
	if _, err := c.bot.Send(mConf); err != nil {
		log.Println("replyWithFile err:", err.Error())
	}
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
	buffer := &bytes.Buffer{}
	fmt.Fprintln(buffer, "Predefined commands:")
	fmt.Fprintln(buffer, "/help - prints help message")
	fmt.Fprintln(buffer, "/file &lt;PATH&gt; - sends back a file with the given path")
	fmt.Fprintln(buffer)
	fmt.Fprintln(buffer, "User commands (script aliases):")
	for name, script := range c.cfg.Commands {
		fmt.Fprintf(buffer, "%s - <i>%s</i>\n", name, script)
	}

	c.replyWithHTML(toMsg, buffer.String())
}
