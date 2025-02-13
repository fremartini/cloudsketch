package load_balancer

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/list"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *az.Context) ([]*az.Resource, error) {
	clientFactory, err := armnetwork.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewLoadBalancersClient()

	lb, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	targets, err := getBackendTargets(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	dependsOn = append(dependsOn, targets...)

	resource := &az.Resource{
		Id:            *lb.ID,
		Name:          *lb.Name,
		Type:          *lb.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     dependsOn,
	}

	resources := []*az.Resource{resource}

	frontends, err := getFrontends(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	resources = append(resources, frontends...)

	return resources, nil
}

func getFrontends(clientFactory *armnetwork.ClientFactory, ctx *az.Context) ([]*az.Resource, error) {
	client := clientFactory.NewLoadBalancerFrontendIPConfigurationsClient()

	pager := client.NewListPager(ctx.ResourceGroup, ctx.ResourceName, nil)

	var nics []*armnetwork.FrontendIPConfiguration
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.LoadBalancerFrontendIPConfigurationListResult.Value != nil {
			nics = append(nics, resp.LoadBalancerFrontendIPConfigurationListResult.Value...)
		}
	}

	return list.Map(nics, func(nic *armnetwork.FrontendIPConfiguration) *az.Resource {
		dependsOn := []string{ctx.ResourceId}

		subnet := strings.ToLower(*nic.Properties.Subnet.ID)

		dependsOn = append(dependsOn, subnet)

		return &az.Resource{
			Id:            *nic.ID,
			Name:          *nic.Name,
			Type:          *nic.Type,
			ResourceGroup: ctx.ResourceGroup,
			DependsOn:     dependsOn,
		}
	}), nil
}

func getBackendTargets(clientFactory *armnetwork.ClientFactory, ctx *az.Context) ([]string, error) {
	client := clientFactory.NewLoadBalancerNetworkInterfacesClient()

	pager := client.NewListPager(ctx.ResourceGroup, ctx.ResourceName, nil)

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

	return list.Map(nics, func(nic *armnetwork.Interface) string {
		return *nic.ID
	}), nil
}
