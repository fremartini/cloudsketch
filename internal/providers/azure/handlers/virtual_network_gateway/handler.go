package virtual_network_gateway

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

	client := clientFactory.NewVirtualNetworkGatewaysClient()

	virtualNetworkGateway, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}
	for _, ipconfiguration := range virtualNetworkGateway.Properties.IPConfigurations {
		if ipconfiguration.Properties.PublicIPAddress != nil {
			dependsOn = append(dependsOn, *ipconfiguration.Properties.PublicIPAddress.ID)
		}

		if ipconfiguration.Properties.Subnet != nil {
			dependsOn = append(dependsOn, *ipconfiguration.Properties.Subnet.ID)
		}
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *virtualNetworkGateway.Type,
		DependsOn: dependsOn,
	}

	resources := []*models.Resource{resource}

	connections, err := getConnections(clientFactory, ctx, ctx.ResourceId)

	if err != nil {
		return nil, err
	}

	resources = append(resources, connections...)

	return resources, nil
}

func getConnections(clientFactory *armnetwork.ClientFactory, ctx *azContext.Context, virtualNetworkGatewayId string) ([]*models.Resource, error) {
	connectionsClient := clientFactory.NewVirtualNetworkGatewayConnectionsClient()

	pager := connectionsClient.NewListPager(ctx.ResourceGroupName, nil)

	var connections []*armnetwork.VirtualNetworkGatewayConnection
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.VirtualNetworkGatewayConnectionListResult.Value != nil {
			connections = append(connections, resp.VirtualNetworkGatewayConnectionListResult.Value...)
		}
	}

	models := list.Map(connections, func(connection *armnetwork.VirtualNetworkGatewayConnection) *models.Resource {
		return &models.Resource{
			Id:   *connection.ID,
			Name: *connection.Name,
			Type: *connection.Type,
			DependsOn: []string{
				virtualNetworkGatewayId,
				*connection.Properties.Peer.ID,
			},
		}
	})

	return models, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
