package config

import (
	_ "embed"
	"encoding/json"
)

//go:embed config.json
var configFile []byte

type config struct {
	Version        string `json:"version"`
	SentryDsn      string `json:"sentry-dsn"`
	GenServiceUrl  string `json:"gen-service-url"`
	GitHubClientId string `json:"github-client-id"`
	CLIUpdaterURL  string `json:"cli-updater-url"`
}

var Config config

func Load() error {
	var c config
	err := json.Unmarshal(configFile, &c)

	if err != nil {
		return err
	}

	Config = c
	return nil
}
