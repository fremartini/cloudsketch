package virtual_network

import (
	"azsample/internal/az"
	"azsample/internal/list"
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

	client := clientFactory.NewSubnetsClient()

	pager := client.NewListPager(ctx.ResourceGroup, ctx.ResourceName, nil)

	var subnets []*armnetwork.Subnet
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		subnets = append(subnets, resp.Value...)
	}

	vnet := &az.Resource{
		Id:            ctx.ResourceId,
		Name:          ctx.ResourceName,
		Type:          az.VIRTUAL_NETWORK,
		ResourceGroup: ctx.ResourceGroup,
	}

	resources := list.Map(subnets, func(subnet *armnetwork.Subnet) *az.Resource {
		snet := &az.Resource{
			Id:            *subnet.ID,
			Name:          *subnet.Name,
			Type:          *subnet.Type,
			ResourceGroup: ctx.ResourceGroup,
			DependsOn:     []string{vnet.Id},
		}

		return snet
	})

	resources = append(resources, vnet)

	return resources, nil
}
