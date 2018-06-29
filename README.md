# TelegramRemote

TelegramRemote is an application wich allows you to execute commands on your PC remotely via Telegram.

How to install and configure:
* [Download](https://github.com/unkeep/TelegramRemote/releases/download/v1.0/TelegramRemote.zip) the TelegramRemote application for windows or build it yourself from source code using `go build`.
* [Create](https://core.telegram.org/bots#3-how-do-i-create-a-bot) telegram bot. Name it like 'MyLaptop_bot' for example.
* Put your bot token in to the `config.json` file.
* Put your telegram user name (without '@') int to the `config.json`.
* Optional. You can add aliases to the most used commands into the `config.json`. `/help` - will list all your configured aliases.
* Launch TelegramRemote.exe
