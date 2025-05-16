package config

import (
	"cloudsketch/internal/marshall"
	"os"
	"path"
	"path/filepath"
)

const (
	CONFIG_FILE = ".cloudsketch.json"
)

type config struct {
	Blacklist []string
}

func Read() (*config, bool) {
	executable, err := os.Executable()

	if err != nil {
		panic(err)
	}

	executablePath := filepath.Dir(executable)
	configFilePath := path.Join(executablePath, CONFIG_FILE)

	return marshall.UnmarshalIfExists[config](configFilePath)
}
