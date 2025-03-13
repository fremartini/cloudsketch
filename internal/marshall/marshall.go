package marshall

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

func UnmarshalIfExists[T any](file string) (*T, bool) {
	if !FileExists(file) {
		return nil, false
	}

	config, err := UnmarshallResources[T](file)

	if err != nil {
		panic(err)
	}

	return config, true
}

func FileExists(file string) bool {
	_, err := os.Stat(file)

	return !errors.Is(err, os.ErrNotExist)
}

func MarshallResources[T any](file string, resources T) error {
	bytes, err := json.MarshalIndent(resources, "", "\t")

	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(file)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.Write(bytes)

	return err
}

func UnmarshallResources[T any](file string) (*T, error) {
	bytes, err := os.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	var config *T

	err = json.Unmarshal(bytes, &config)

	return config, err
}
