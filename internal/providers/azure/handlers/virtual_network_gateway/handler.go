package virtual_network_gateway

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
	clientFactory, err := armnetwork.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewVirtualNetworkGatewaysClient()

	virtualNetworkGateway, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}
	for _, ipconfiguration := range virtualNetworkGateway.Properties.IPConfigurations {
		if ipconfiguration.Properties.PublicIPAddress != nil {
			dependsOn = append(dependsOn, *ipconfiguration.Properties.PublicIPAddress.ID)
		}

		if ipconfiguration.Properties.Subnet != nil {
			dependsOn = append(dependsOn, *ipconfiguration.Properties.Subnet.ID)
		}
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *virtualNetworkGateway.Type,
		DependsOn: dependsOn,
	}

	resources := []*models.Resource{resource}

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
