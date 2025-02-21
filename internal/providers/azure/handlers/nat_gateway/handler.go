package nat_gateway

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armnetwork.NewNatGatewaysClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	ngw, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	properties := map[string]string{}
	dependsOn := []string{}

	for _, subnet := range ngw.Properties.Subnets {
		dependsOn = append(dependsOn, *subnet.ID)
	}

	resource := &models.Resource{
		Id:         *ngw.ID,
		Name:       *ngw.Name,
		Type:       *ngw.Type,
		DependsOn:  dependsOn,
		Properties: properties,
	}

	return []*models.Resource{resource}, nil
}
