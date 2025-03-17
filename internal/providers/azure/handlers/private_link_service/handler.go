package private_link_service

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
	clientFactory, err := armnetwork.NewPrivateLinkServicesClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pls, err := clientFactory.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	pls_target := pls.Properties.LoadBalancerFrontendIPConfigurations[0].ID

	resource := &models.Resource{
		Id:        *pls.ID,
		Name:      *pls.Name,
		Type:      *pls.Type,
		DependsOn: []string{*pls_target},
	}

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
