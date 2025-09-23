package postgres_flexible_server

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers/v5"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armpostgresqlflexibleservers.NewServersClient(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pfsql, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	if pfsql.Properties.Network.DelegatedSubnetResourceID != nil {
		dependsOn = append(dependsOn, strings.ToLower(*pfsql.Properties.Network.DelegatedSubnetResourceID))
	}

	if pfsql.Properties.Network.PrivateDNSZoneArmResourceID != nil {
		dependsOn = append(dependsOn, strings.ToLower(*pfsql.Properties.Network.PrivateDNSZoneArmResourceID))
	}

	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return []*models.Resource{}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
