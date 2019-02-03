package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type task struct {
	cmd    *exec.Cmd
	cmdStr string
	cancel context.CancelFunc
}

type taskMap map[int]*task

type controller struct {
	bot        *tgbotapi.BotAPI
	botUpdates tgbotapi.UpdatesChannel
	cfg        *config
	handlers   []handlerFunc
	workingDir string
	tasks      taskMap
	taskdID    int
	tasksMutex sync.Mutex
}

func createController(bot *tgbotapi.BotAPI, cfg *config) (*controller, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updateChan, err := bot.GetUpdatesChan(u)

	if err != nil {
		return nil, err
	}

	c := &controller{
		bot:        bot,
		botUpdates: updateChan,
		cfg:        cfg,
		tasks:      taskMap{},
	}

	c.handlers = []handlerFunc{
		c.msgCheckHandler,
		c.authHandler,
		c.helpHandler,
		c.fileHandler,
		c.cdHandler,
		c.tasksHandler,
		c.killHandler,
		c.cmdHandler,
		c.unsupportedHandler,
	}

	return c, nil
}

type handlerFunc func(tgbotapi.Update) bool

func (c *controller) start() {
	log.Println("Handling chat updates...")

	for update := range c.botUpdates {
		for _, handler := range c.handlers {
			if handler(update) {
				break
			}
		}
	}
}

func (c *controller) msgCheckHandler(update tgbotapi.Update) bool {
	return update.Message == nil
}

func (c *controller) authHandler(update tgbotapi.Update) bool {
	if !c.isAuthorizedUser(update.Message.From.UserName) {
		c.replyWithText(update.Message, "Sorry, you are not authorized")
		return true
	}
	return false
}

func (c *controller) helpHandler(update tgbotapi.Update) bool {
	if update.Message.Text == "/help" {
		c.replyToHelp(update.Message)
		return true
	}
	return false
}

func (c *controller) fileHandler(update tgbotapi.Update) bool {
	if !strings.HasPrefix(update.Message.Text, "/file ") {
		return false
	}

	filePath := path.Join(c.workingDir, strings.TrimPrefix(update.Message.Text, "/file "))
	c.replyWithFile(update.Message, filePath)
	return true
}

func (c *controller) cdHandler(update tgbotapi.Update) bool {
	if !strings.HasPrefix(update.Message.Text, "/cd ") {
		return false
	}

	dirPath := strings.TrimPrefix(update.Message.Text, "/cd ")
	if fInfo, err := os.Stat(dirPath); err == nil && fInfo.IsDir() {
		c.workingDir = dirPath
		c.replyWithText(update.Message, fmt.Sprintf("Current dir is set to '%s'", dirPath))
	} else {
		c.replyWithText(update.Message, fmt.Sprintf("Invalid dir path"))
	}

	return true
}

func (c *controller) tasksHandler(update tgbotapi.Update) bool {
	if !strings.HasPrefix(update.Message.Text, "/tasks") {
		return false
	}

	c.tasksMutex.Lock()
	var reply string
	for id, t := range c.tasks {
		reply += fmt.Sprintf("%d - %s\n", id, t.cmdStr)
	}
	c.tasksMutex.Unlock()

	if reply == "" {
		reply = "no running tasks"
	}

	c.replyWithText(update.Message, reply)

	return true
}

func (c *controller) killHandler(update tgbotapi.Update) bool {
	if !strings.HasPrefix(update.Message.Text, "/kill ") {
		return false
	}

	idStr := strings.TrimPrefix(update.Message.Text, "/kill ")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.replyWithText(update.Message, "invalid task id: "+idStr)
		return true
	}

	if t, ok := c.tasks[id]; ok {
		t.cancel()
	} else {
		c.replyWithText(update.Message, fmt.Sprintf("task with ID %d not found", id))
	}

	return true
}

func (c *controller) cmdHandler(update tgbotapi.Update) bool {
	var commandStr string
	if strings.HasPrefix(update.Message.Text, "/") {
		if userCommand, ok := c.cfg.Commands[update.Message.Text]; ok {
			commandStr = userCommand
		} else {
			return false
		}
	} else {
		commandStr = update.Message.Text
	}

	go c.runTask(commandStr, update.Message)
	return true
}

func (c *controller) unsupportedHandler(update tgbotapi.Update) bool {
	c.replyWithText(update.Message, "unsupported command")
	return true
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
	if _, err := c.bot.Send(mConf); err != nil {
		log.Println("replyWithFile err:", err.Error())
		c.replyWithText(toMsg, err.Error())
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

func (c *controller) runTask(command string, msg *tgbotapi.Message) {
	log.Printf("executing '%s'", command)

	words := strings.Split(command, " ")
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, words[0], words[1:]...)
	cmd.Dir = c.workingDir
	var out bytes.Buffer
	cmd.Stdout = &out

	c.tasksMutex.Lock()
	c.taskdID++
	id := c.taskdID
	c.tasks[id] = &task{
		cmd:    cmd,
		cmdStr: command,
		cancel: cancel,
	}
	c.tasksMutex.Unlock()

	defer func() {
		c.tasksMutex.Lock()
		delete(c.tasks, id)
		c.tasksMutex.Unlock()
	}()

	if err := cmd.Run(); err != nil {
		log.Println("cmd.Run err:", err.Error())
		c.replyWithText(msg, err.Error())
		return
	}

	output := out.String()
	log.Println("Out:", output)
	if output != "" {
		c.replyWithText(msg, output)
		return
	}
}
