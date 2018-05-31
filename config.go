package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"errors"

	"path/filepath"
)

// Data abstract containing methods for saving and loading.

// Data an abstract struct used for it's functions to save and load config files.
type Data struct{}

func (d *Data) save(saveLoc string, inter interface{}) error {
	// Make all the directories
	if err := os.MkdirAll(filepath.Dir(saveLoc), os.ModeDir|0775); err != nil {
		return err
	}

	data, err := json.MarshalIndent(inter, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(saveLoc, data, 0660)
}

func (d *Data) load(saveLoc string, inter interface{}) error {

	if _, err := os.Stat(saveLoc); os.IsNotExist(err) {
		return DefaultConfigSavedError
	} else if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(saveLoc)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, inter); err != nil {
		return err
	}

	return nil

}

// ========== Main configuration.

// ConfigSaveLocation the location to save the config to.
var ConfigSaveLocation = "config.json"

// DefaultConfigSavedError an error returned if the default config is saved.
var DefaultConfigSavedError = errors.New("the default config has been saved, please edit it")

// DefaultConfig the default configuration to save.
var DefaultConfig = Config{
	Data:              Data{},
	Discord: DiscordConfig{"TOKEN"},
	Database: DatabaseConfig{"Databaseuri"},
}

// Config the main configuration.
type Config struct {
	Data              `json:"-"`
	Discord DiscordConfig `json:"discord"`
	Database DatabaseConfig `json:"database"`
	Redis RedisConfig `json:"redis"`
}

type RedisConfig struct {
	// TODO
}

type DatabaseConfig struct {
	URI string `json:"uri"`
}

type DiscordConfig struct {
	Token string `json:"token"`
}

// Save saves the config.
func (c *Config) Save() error {
	saveLoc, envThere := os.LookupEnv("CONFIG_LOC")
	if !envThere {
		saveLoc = ConfigSaveLocation
	}

	return c.save(saveLoc, c)
}

// Load loads the config.
func (c *Config) Load() error {

	saveLoc, envThere := os.LookupEnv("CONFIG_LOC")
	if !envThere {
		saveLoc = ConfigSaveLocation
	}

	if err := c.load(saveLoc, c); err == DefaultConfigSavedError {
		if err := DefaultConfig.Save(); err != nil {
			return err
		}
		return DefaultConfigSavedError
	} else if err != nil {
		return err
	}

	return nil

}