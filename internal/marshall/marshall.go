package marshall

import (
	"cloudsketch/internal/drawio/models"
	"encoding/json"
	"errors"
	"log"
	"os"
)

func UnmarshalIfExists(file string) ([]*models.Resource, bool) {
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

func MarshallResources(file string, resources []*models.Resource) error {
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

func UnmarshallResources(file string) ([]*models.Resource, error) {
	bytes, err := os.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	var resources []*models.Resource

	err = json.Unmarshal(bytes, &resources)

	return resources, err
}
