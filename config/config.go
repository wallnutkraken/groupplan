// Package config manages the JSON configuration for groupplan
package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

// ConfigPath is the local path to where the config file should be
const ConfigPath = "groupplan.config.json"

// AppSettings contains the application settings
type AppSettings struct {
	Hostname        string
	DiscordKey      string
	DiscordSecret   string
	ECDSAPrivateKey string
}

// GetDefault returns the default settings object with a generate private key
func GetDefault() AppSettings {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic("could not create an ecdsa private key: " + err.Error())
	}
	// Encode to x509
	encoded, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		panic("could not marshal to x509: " + err.Error())
	}

	pemEncoded := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: encoded,
	})
	return AppSettings{
		ECDSAPrivateKey: string(pemEncoded),
	}
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
