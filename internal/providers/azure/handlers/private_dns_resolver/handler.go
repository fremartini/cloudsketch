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

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armdnsresolver.NewClientFactory(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewDNSResolversClient()

	privateDnsZone, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{*privateDnsZone.Properties.VirtualNetwork.ID}

	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return []*models.Resource{}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
