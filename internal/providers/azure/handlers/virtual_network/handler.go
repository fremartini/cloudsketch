package virtual_network

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"
	"strings"

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

	client := clientFactory.NewVirtualNetworksClient()

	vnet, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	vnetResource, err := mapVirtualNetworkResource(&vnet, ctx)

	if err != nil {
		return nil, err
	}

	// subnets are a subresource of virtual networks so they must be fetched together
	subnets, err := mapSubnetResources(vnet.Properties.Subnets, vnetResource.Id)

	if err != nil {
		return nil, err
	}

	return append(subnets, vnetResource), nil
}

func mapVirtualNetworkResource(vnet *armnetwork.VirtualNetworksClientGetResponse, ctx *azContext.Context) (*models.Resource, error) {
	addressPrefixes := vnet.Properties.AddressSpace.AddressPrefixes

	properties := map[string][]string{}

	// virtual networks can have multiple address ranges. If this is the case hide the size
	if len(addressPrefixes) == 1 {
		addressPrefix := strings.Split(*addressPrefixes[0], "/")[1]

		properties["size"] = []string{addressPrefix}
	}

	resource := &models.Resource{
		Id:         ctx.ResourceId,
		Name:       ctx.ResourceName,
		Type:       types.VIRTUAL_NETWORK,
		Properties: properties,
	}

	return resource, nil
}

func mapSubnetResources(subnets []*armnetwork.Subnet, vnetId string) ([]*models.Resource, error) {
	resources := list.Map(subnets, func(subnet *armnetwork.Subnet) *models.Resource {
		dependsOn := []string{vnetId}

		routeTable := subnet.Properties.RouteTable

		if routeTable != nil {
			dependsOn = append(dependsOn, strings.ToLower(*routeTable.ID))
		}

		nsg := subnet.Properties.NetworkSecurityGroup

		if nsg != nil {
			dependsOn = append(dependsOn, strings.ToLower(*nsg.ID))
		}

		addressPrefix := strings.Split(*subnet.Properties.AddressPrefix, "/")[1]

		properties := map[string][]string{
			"size": {addressPrefix},
		}

		snet := &models.Resource{
			Id:         *subnet.ID,
			Name:       *subnet.Name,
			Type:       *subnet.Type,
			DependsOn:  dependsOn,
			Properties: properties,
		}

		return snet
	})

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
