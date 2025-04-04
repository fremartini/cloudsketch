package express_route_gateway

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

	client := clientFactory.NewExpressRouteGatewaysClient()

	gateway, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	vhub := gateway.Properties.VirtualHub.ID

	properties := map[string]any{
		"peerings": list.Map(gateway.Properties.ExpressRouteConnections, func(connection *armnetwork.ExpressRouteConnection) string {
			return *connection.Properties.ExpressRouteCircuitPeering.ID
		}),
	}

	resource := &models.Resource{
		Id:         ctx.ResourceId,
		Name:       ctx.ResourceName,
		Type:       *gateway.Type,
		DependsOn:  []string{*vhub},
		Properties: properties,
	}

	resources := []*models.Resource{resource}

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
