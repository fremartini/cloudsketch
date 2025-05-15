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

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armnetwork.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewLoadBalancersClient()

	lb, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	resource := &models.Resource{
		Id:        *lb.ID,
		Name:      *lb.Name,
		Type:      *lb.Type,
		DependsOn: []string{},
	}

	resources := []*models.Resource{resource}

	backendPools, err := getBackendPools(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	frontends, err := getFrontends(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	//nics, err := getNicsForPool(clientFactory, ctx)

	resources = append(resources, backendPools...)
	resources = append(resources, frontends...)

	return resources, nil
}

func getFrontends(clientFactory *armnetwork.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewLoadBalancerFrontendIPConfigurationsClient()

	pager := client.NewListPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

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
		dependsOn := []string{ctx.ResourceId}

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

	pager := client.NewListPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

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
		dependsOn := []string{ctx.ResourceId}

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

func getNicsForPool(clientFactory *armnetwork.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewLoadBalancerNetworkInterfacesClient()

	pager := client.NewListPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

	var nics []*armnetwork.Interface
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.InterfaceListResult.Value != nil {
			nics = append(nics, resp.InterfaceListResult.Value...)
		}
	}

	return list.Map(nics, func(nic *armnetwork.Interface) *models.Resource {
		return &models.Resource{
			Id:        *nic.ID,
			Name:      *nic.Name,
			Type:      *nic.Type,
			DependsOn: []string{},
		}
	}), nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
