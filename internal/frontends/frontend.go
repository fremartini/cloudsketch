package frontends

import "cloudsketch/internal/frontends/models"

type Frontend interface {
	WriteDiagram(resources []*models.Resource, filename string) error
}
