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

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armnetwork.NewClientFactory(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewExpressRouteGatewaysClient()

	gateway, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	vhub := gateway.Properties.VirtualHub.ID

	properties := map[string][]string{
		"peerings": list.Map(gateway.Properties.ExpressRouteConnections, func(connection *armnetwork.ExpressRouteConnection) string {
			return *connection.Properties.ExpressRouteCircuitPeering.ID
		}),
	}

	resource.DependsOn = append(resource.DependsOn, *vhub)
	resource.Properties = properties

	return []*models.Resource{}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
