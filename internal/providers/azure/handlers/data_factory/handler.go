package data_factory

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v9"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armdatafactory.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewFactoriesClient()

	adf, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	resource := &models.Resource{
		Id:   *adf.ID,
		Name: *adf.Name,
		Type: *adf.Type,
	}

	resources := []*models.Resource{resource}

	integration_runtimes, err := getIntegrationRuntimes(clientFactory, ctx, adf.ID)

	if err != nil {
		return nil, err
	}

	resources = append(resources, integration_runtimes...)

	networks, err := getManagedVirtualNetworks(clientFactory, ctx, adf.Name)

	if err != nil {
		return nil, err
	}

	endpoints, err := getManagedPrivateEndpoints(clientFactory, ctx, adf.ID, networks[0].Name)

	if err != nil {
		return nil, err
	}

	resources = append(resources, endpoints...)

	return resources, nil
}

func getManagedVirtualNetworks(clientFactory *armdatafactory.ClientFactory, ctx *azContext.Context, adfName *string) ([]*armdatafactory.ManagedVirtualNetworkResource, error) {
	client := clientFactory.NewManagedVirtualNetworksClient()

	pager := client.NewListByFactoryPager(ctx.ResourceGroupName, *adfName, nil)

	var networks []*armdatafactory.ManagedVirtualNetworkResource
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.ManagedVirtualNetworkListResponse.Value != nil {
			networks = append(networks, resp.ManagedVirtualNetworkListResponse.Value...)
		}
	}

	return networks, nil
}

func getManagedPrivateEndpoints(clientFactory *armdatafactory.ClientFactory, ctx *azContext.Context, adfId, managedVirtualNetworkName *string) ([]*models.Resource, error) {
	client := clientFactory.NewManagedPrivateEndpointsClient()

	pager := client.NewListByFactoryPager(ctx.ResourceGroupName, ctx.ResourceName, *managedVirtualNetworkName, nil)

	var endpoints []*armdatafactory.ManagedPrivateEndpointResource
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.ManagedPrivateEndpointListResponse.Value != nil {
			endpoints = append(endpoints, resp.ManagedPrivateEndpointListResponse.Value...)
		}
	}

	resources := list.Map(endpoints, func(endpoint *armdatafactory.ManagedPrivateEndpointResource) *models.Resource {
		return &models.Resource{
			Id:        *endpoint.ID,
			Name:      *endpoint.Name,
			Type:      *endpoint.Type,
			DependsOn: []string{*adfId},
		}
	})

	return resources, nil
}

func getIntegrationRuntimes(clientFactory *armdatafactory.ClientFactory, ctx *azContext.Context, adfId *string) ([]*models.Resource, error) {
	client := clientFactory.NewIntegrationRuntimesClient()

	pager := client.NewListByFactoryPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

	var integration_runtimes []*armdatafactory.IntegrationRuntimeResource
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.IntegrationRuntimeListResponse.Value != nil {
			integration_runtimes = append(integration_runtimes, resp.IntegrationRuntimeListResponse.Value...)
		}
	}

	resources := list.Map(integration_runtimes, func(ir *armdatafactory.IntegrationRuntimeResource) *models.Resource {
		return &models.Resource{
			Id:        *ir.ID,
			Name:      *ir.Name,
			Type:      *ir.Type,
			DependsOn: []string{*adfId},
		}
	})

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
