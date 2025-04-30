package providers

import "cloudsketch/internal/frontends/drawio/models"

type Provider interface {
	FetchResources(subscriptionId string) ([]*models.Resource, string, error)
}
