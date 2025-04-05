package private_dns_resolver

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armdnsresolver.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewDNSResolversClient()

	privateDnsZone, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *privateDnsZone.Type,
		DependsOn: []string{*privateDnsZone.Properties.VirtualNetwork.ID},
	}

	resources := []*models.Resource{resource}

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
