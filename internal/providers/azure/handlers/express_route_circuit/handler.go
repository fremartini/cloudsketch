package express_route_circuit

import (
	"cloudsketch/internal/list"
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

	client := clientFactory.NewExpressRouteCircuitsClient()

	circuit, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	properties := map[string][]string{
		"peerings": list.Map(circuit.Properties.Peerings, func(peering *armnetwork.ExpressRouteCircuitPeering) string {
			return *peering.ID
		}),
	}

	resource := &models.Resource{
		Id:         ctx.ResourceId,
		Name:       ctx.ResourceName,
		Type:       *circuit.Type,
		DependsOn:  []string{},
		Properties: properties,
	}

	resources := []*models.Resource{resource}

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
