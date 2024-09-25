package settings

import (
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

const settingsFile string = "settings.yaml"

type AppSettings struct {
	Debug bool `yaml:"debug"`
	Bot   Bot  `yaml:"bot"`
	Db    Db   `yaml:"db"`
}

type Bot struct {
	Token          string `yaml:"token"`
	UsePolling     bool   `yaml:"usePolling"`
	WebhookBaseUrl string `yaml:"webhookBaseUrl"`
}

type Db struct {
	ConnectionString string `yaml:"connectionString"`
}

func GetSettings() AppSettings {
	settings, err := getSettingsFromFile()
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
	return settings
}

func getSettingsFromFile() (AppSettings, error) {
	settings := AppSettings{}

	if _, err := os.Stat(settingsFile); err != nil {
		return settings, fmt.Errorf("can't read settings file")
	}

	fileContent, err := os.ReadFile(settingsFile)

	if err != nil {
		return settings, err
	}

	if err = yaml.Unmarshal(fileContent, &settings); err != nil {
		return settings, err
	}

	return settings, nil
}
