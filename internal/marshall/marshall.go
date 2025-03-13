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

	r, err := UnmarshallResources[T](file)

	if err != nil {
		panic(err)
	}

	return r, true
}

func FileExists(file string) bool {
	_, err := os.Stat(file)

	return !errors.Is(err, os.ErrNotExist)
}

func MarshallResources[T any](file string, r T) error {
	bytes, err := json.MarshalIndent(r, "", "\t")

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

	var r *T

	err = json.Unmarshal(bytes, &r)

	return r, err
}
