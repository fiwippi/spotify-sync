package client

import (
	"encoding/json"
	"os"
)

// Global config var to be passed around
var details *Config

// Config details used to connect to the server which are saved to become persistent
type Config struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Address   string `json:"address"`
	ServerKey string `json:"server_key"`
	AdminKey  string `json:"admin_key"`
}

// Saves the config file
func saveConfig(c *Config) error {
	file, err := os.OpenFile("config.json", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(c)
}

// Loads the config file
func openConfig() (*Config, error) {
	file, err := os.OpenFile("config.json", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	jsonDecoder := json.NewDecoder(file)
	err = jsonDecoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
