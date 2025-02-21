package public_ip_address

import (
	"context"

	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armnetwork.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewPublicIPAddressesClient()

	pip, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	if pip.Properties.NatGateway != nil {
		pip_target := pip.Properties.NatGateway.ID
		dependsOn = append(dependsOn, *pip_target)
	}

	resource := &models.Resource{
		Id:        *pip.ID,
		Name:      *pip.Name,
		Type:      *pip.Type,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
}
