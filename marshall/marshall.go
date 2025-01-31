package marshall

import (
	"azsample/internal/az"
	"encoding/json"
	"errors"
	"log"
	"os"
)

func UnmarshalIfExists(file string) (bool, []*az.Resource) {
	if !FileExists(file) {
		return false, nil
	}

	resources := UnmarshallResources(file)

	return true, resources
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

func UnmarshallResources(file string) []*az.Resource {
	bytes, err := os.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	var resources []*az.Resource

	json.Unmarshal(bytes, &resources)

	return resources
}
