package main

import (
	"fmt"
	"os"

	"github.com/wallnutkraken/groupplan/config"
	"github.com/wallnutkraken/groupplan/groupdata"
	"github.com/wallnutkraken/groupplan/httpend"
	"github.com/wallnutkraken/groupplan/userman"
)

// DBPath is the static path to the database file
var DBPath = "groupplan.sqlite3"

func main() {
	// Load the config
	cfg, err := config.Load()
	if err != nil {
		// Couldn't read the config, write a new one and inform the user
		if err := config.GetDefault().Save(); err != nil {
			// Failed to write, inform the user and exit
			fmt.Printf("Failed to create a config file at [%s], check your file permissions", config.ConfigPath)
			os.Exit(1)
		}
		fmt.Printf("Failed reading the config at [%s], new one created, please fill it out and launch the application again\n", config.ConfigPath)
		os.Exit(1)
	}
	// Spin up the database, locally
	db, err := groupdata.New(DBPath)
	if err != nil {
		fmt.Printf("Failed loading database at [%s]: %s", DBPath, err.Error())
	}
	endpoint := httpend.New(cfg, userman.New(db.Users()))
	// Start listening
	if err := endpoint.Start(); err != nil {
		fmt.Printf("Error while listening: %s\n", err.Error())
		os.Exit(1)
	}
}
