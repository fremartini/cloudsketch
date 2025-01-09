package guid

import (
	"strings"

	"github.com/google/uuid"
)

func NewGuidAlphanumeric() string {
	id := uuid.New()

	return strings.ReplaceAll(id.String(), "-", "")
}
