// Package config manages the JSON configuration for groupplan
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// ConfigPath is the local path to where the config file should be
const ConfigPath = "groupplan.config.json"

// AppSettings contains the application settings
type AppSettings struct {
	Hostname      string
	DiscordKey    string
	DiscordSecret string
}

// GetDefault returns the default settings object
func GetDefault() AppSettings {
	return AppSettings{}
}

// Load loads the settings file (from current working directory)
func Load() (AppSettings, error) {
	data, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		return AppSettings{}, fmt.Errorf("failed reading config file: %w", err)
	}
	// Unmarshal the JSON
	settings := AppSettings{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, fmt.Errorf("failed unmarshalling settings struct: %w", err)
	}
	return settings, nil
}

// Save writes the settings to disk
func (a AppSettings) Save() error {
	// JSON-encode the settings
	data, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return fmt.Errorf("failed marshalling app settings: %w", err)
	}
	return ioutil.WriteFile(ConfigPath, data, os.ModePerm)
}
