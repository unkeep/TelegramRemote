package main

import (
	"encoding/json"
	"os"
)

const defaultConfigPath string = "config.json"

type config struct {
	BotToken  string            `json:"botToken"`
	WhiteList []string          `json:"whiteList"`
	Commands  map[string]string `json:"commands"`
}

func loadConfig() (*config, error) {
	path := defaultConfigPath
	appArgs := os.Args
	if len(appArgs) > 1 {
		path = os.Args[1]
	}

	configFile, err := os.Open(path)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(configFile)

	c := new(config)
	if err := jsonParser.Decode(c); err != nil {
		return nil, err
	}

	return c, nil
}
