package config

import "cloudsketch/internal/marshall"

const (
	CONFIG_FILE = ".cloudsketch.json"
)

type config struct {
	Blacklist []string
}

func Read() (*config, bool) {
	return marshall.UnmarshalIfExists[config](CONFIG_FILE)
}
