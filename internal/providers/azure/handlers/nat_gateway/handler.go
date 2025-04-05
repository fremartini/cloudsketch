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

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armnetwork.NewNatGatewaysClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	ngw, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	for _, subnet := range ngw.Properties.Subnets {
		dependsOn = append(dependsOn, *subnet.ID)
	}

	for _, pip := range ngw.Properties.PublicIPAddresses {
		dependsOn = append(dependsOn, *pip.ID)
	}

	resource := &models.Resource{
		Id:        *ngw.ID,
		Name:      *ngw.Name,
		Type:      *ngw.Type,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
