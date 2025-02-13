package marshall

import (
	"cloudsketch/internal/az"
	"encoding/json"
	"errors"
	"log"
	"os"
)

func UnmarshalIfExists(file string) ([]*az.Resource, bool) {
	if !FileExists(file) {
		return nil, false
	}

	resources, err := UnmarshallResources(file)

	if err != nil {
		panic(err)
	}

	return resources, true
}

func FileExists(file string) bool {
	_, err := os.Stat(file)

	return !errors.Is(err, os.ErrNotExist)
}

func MarshallResources(file string, resources []*az.Resource) error {
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

func UnmarshallResources(file string) ([]*az.Resource, error) {
	bytes, err := os.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	var resources []*az.Resource

	err = json.Unmarshal(bytes, &resources)

	return resources, err
}
