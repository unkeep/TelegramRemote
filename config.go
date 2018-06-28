package main

import (
	"encoding/json"
	"os"
)

const configPath string = "config.json"

type config struct {
	BotToken string            `json:"botToken"`
	Commands map[string]string `json:"commands"`
}

func loadConfig() (*config, error) {
	configFile, err := os.Open(configPath)
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
