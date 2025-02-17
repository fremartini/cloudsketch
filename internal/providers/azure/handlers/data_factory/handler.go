package data_factory

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v9"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *azContext.Context) ([]*az.Resource, error) {
	clientFactory, err := armdatafactory.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewFactoriesClient()

	adf, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	resource := &az.Resource{
		Id:            *adf.ID,
		Name:          *adf.Name,
		Type:          *adf.Type,
		ResourceGroup: ctx.ResourceGroup,
	}

	resources := []*az.Resource{resource}

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

	pager := client.NewListByFactoryPager(ctx.ResourceGroup, *adfName, nil)

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

func getManagedPrivateEndpoints(clientFactory *armdatafactory.ClientFactory, ctx *azContext.Context, adfId, managedVirtualNetworkName *string) ([]*az.Resource, error) {
	client := clientFactory.NewManagedPrivateEndpointsClient()

	pager := client.NewListByFactoryPager(ctx.ResourceGroup, ctx.ResourceName, *managedVirtualNetworkName, nil)

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

	resources := list.Map(endpoints, func(endpoint *armdatafactory.ManagedPrivateEndpointResource) *az.Resource {
		return &az.Resource{
			Id:            *endpoint.ID,
			Name:          *endpoint.Name,
			Type:          *endpoint.Type,
			ResourceGroup: ctx.ResourceGroup,
			DependsOn:     []string{*adfId},
		}
	})

	return resources, nil
}

func getIntegrationRuntimes(clientFactory *armdatafactory.ClientFactory, ctx *azContext.Context, adfId *string) ([]*az.Resource, error) {
	client := clientFactory.NewIntegrationRuntimesClient()

	pager := client.NewListByFactoryPager(ctx.ResourceGroup, ctx.ResourceName, nil)

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

	resources := list.Map(integration_runtimes, func(ir *armdatafactory.IntegrationRuntimeResource) *az.Resource {
		return &az.Resource{
			Id:            *ir.ID,
			Name:          *ir.Name,
			Type:          *ir.Type,
			ResourceGroup: ctx.ResourceGroup,
			DependsOn:     []string{*adfId},
		}
	})

	return resources, nil
}
