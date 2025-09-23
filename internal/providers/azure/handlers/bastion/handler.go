package bastion

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

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armnetwork.NewBastionHostsClient(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	bastion, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	for _, config := range bastion.Properties.IPConfigurations {
		dependsOn = append(dependsOn, *config.Properties.PublicIPAddress.ID)
		dependsOn = append(dependsOn, *config.Properties.Subnet.ID)
	}

	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
