package private_link_service

import (
	"cloudsketch/internal/az"
	azContext "cloudsketch/internal/providers/azure/context"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *azContext.Context) ([]*az.Resource, error) {
	clientFactory, err := armnetwork.NewPrivateLinkServicesClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pls, err := clientFactory.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	pls_target := pls.Properties.LoadBalancerFrontendIPConfigurations[0].ID

	resource := &az.Resource{
		Id:        *pls.ID,
		Name:      *pls.Name,
		Type:      *pls.Type,
		DependsOn: []string{*pls_target},
	}

	return []*az.Resource{resource}, nil
}
