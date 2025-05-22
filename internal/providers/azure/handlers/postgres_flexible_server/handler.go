package postgres_flexible_server

/*
TODO:
Postgres flexible
	- Connect to dns zone
	- Connect to subnet
*/

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

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armpostgresqlflexibleservers.NewServersClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pfsql, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{
		strings.ToLower(*pfsql.Properties.Network.DelegatedSubnetResourceID),
		strings.ToLower(*pfsql.Properties.Network.PrivateDNSZoneArmResourceID),
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *pfsql.Type,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
