package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

// Config is info from config file
type Config struct {
	IP   string
	Port string

	DBUser string
	DBPass string
	DBName string
}

// ReadConfig reads info from config file
func ReadConfig() Config {
	configFile := flag.String("config", "/etc/GSTORE-insert/GSTORE-insert.conf", "Path to config file")
	flag.Parse()
	var configfile = *configFile
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile, " Try with install param")
	}
	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
	return config
}
