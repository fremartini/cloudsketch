package providers

import "cloudsketch/internal/frontends/models"

type Provider interface {
	FetchResources(subscriptionId string) ([]*models.Resource, string, error)
}
