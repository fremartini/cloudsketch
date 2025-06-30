package guid

import (
	"strings"

	"github.com/google/uuid"
)

func NewGuidAlphanumeric() string {
	id := uuid.New()

	return strings.ReplaceAll(id.String(), "-", "")
}

func IsGuid(maybeGuid string) bool {
	err := uuid.Validate(maybeGuid)

	return err == nil
}
