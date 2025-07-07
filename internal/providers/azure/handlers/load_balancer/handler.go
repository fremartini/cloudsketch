package load_balancer

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"
	"strings"

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

	backendPools, err := getBackendPools(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	frontends, err := getFrontends(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	resources := []*models.Resource{}
	resources = append(resources, backendPools...)
	resources = append(resources, frontends...)

	return resources, nil
}

func getFrontends(clientFactory *armnetwork.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewLoadBalancerFrontendIPConfigurationsClient()

	pager := client.NewListPager(ctx.ResourceGroup, ctx.Resource.Name, nil)

	var frontendConfiguration []*armnetwork.FrontendIPConfiguration
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.LoadBalancerFrontendIPConfigurationListResult.Value != nil {
			frontendConfiguration = append(frontendConfiguration, resp.LoadBalancerFrontendIPConfigurationListResult.Value...)
		}
	}

	return list.Map(frontendConfiguration, func(nic *armnetwork.FrontendIPConfiguration) *models.Resource {
		dependsOn := []string{ctx.Resource.Id}

		if nic.Properties.Subnet != nil {
			subnet := strings.ToLower(*nic.Properties.Subnet.ID)

			dependsOn = append(dependsOn, subnet)
		}

		return &models.Resource{
			Id:        *nic.ID,
			Name:      *nic.Name,
			Type:      *nic.Type,
			DependsOn: dependsOn,
		}
	}), nil
}

func getBackendPools(clientFactory *armnetwork.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewLoadBalancerBackendAddressPoolsClient()

	pager := client.NewListPager(ctx.ResourceGroup, ctx.Resource.Name, nil)

	var pools []*armnetwork.BackendAddressPool
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.LoadBalancerBackendAddressPoolListResult.Value != nil {
			pools = append(pools, resp.LoadBalancerBackendAddressPoolListResult.Value...)
		}
	}

	resources := []*models.Resource{}

	backendPoolsResources := list.Map(pools, func(pool *armnetwork.BackendAddressPool) *models.Resource {
		dependsOn := []string{ctx.Resource.Id}

		return &models.Resource{
			Id:        *pool.ID,
			Name:      *pool.Name,
			Type:      *pool.Type,
			DependsOn: dependsOn,
		}
	})

	resources = append(resources, backendPoolsResources...)

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
