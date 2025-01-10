package public_ip_address

import (
	"azsample/internal/az"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *az.Context) ([]*az.Resource, error) {
	clientFactory, err := armnetwork.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewPublicIPAddressesClient()

	pip, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	if pip.Properties.NatGateway != nil {
		pip_target := pip.Properties.NatGateway.ID
		dependsOn = append(dependsOn, *pip_target)
	}

	resource := &az.Resource{
		Id:            *pip.ID,
		Name:          *pip.Name,
		Type:          *pip.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     dependsOn,
	}

	return []*az.Resource{resource}, nil
}
