package frontends

import "cloudsketch/internal/frontends/drawio/models"

type Frontend interface {
	WriteDiagram(resources []*models.Resource, filename string) error
}
