package marshall

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

func UnmarshalIfExists[T any](path string) (*T, bool) {
	if !FileExists(path) {
		return nil, false
	}

	result, err := UnmarshallResources[T](path)

	if err != nil {
		panic(err)
	}

	return result, true
}

func FileExists(file string) bool {
	_, err := os.Stat(file)

	return !errors.Is(err, os.ErrNotExist)
}

func MarshallResources[T any](path string, r T) error {
	bytes, err := json.MarshalIndent(r, "", "\t")

	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(path)

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

	var t *T

	err = json.Unmarshal(bytes, &t)

	return t, err
}
