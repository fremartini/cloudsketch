package virtual_machine_scale_set

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armcompute.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewVirtualMachineScaleSetsClient()

	vmss, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	for identity := range vmss.Identity.UserAssignedIdentities {
		t := strings.ToLower(identity)
		dependsOn = append(dependsOn, t)
	}

	for _, nic := range vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations {
		for _, ipConfiguration := range nic.Properties.IPConfigurations {
			subnet := strings.ToLower(*ipConfiguration.Properties.Subnet.ID)

			dependsOn = append(dependsOn, subnet)

			for _, loadBalancerBackendAddressPool := range ipConfiguration.Properties.LoadBalancerBackendAddressPools {
				poolId := strings.ToLower(*loadBalancerBackendAddressPool.ID)

				dependsOn = append(dependsOn, poolId)
			}
		}
	}

	resources := []*models.Resource{
		{
			Id:        *vmss.ID,
			Name:      *vmss.Name,
			Type:      *vmss.Type,
			DependsOn: dependsOn,
		},
	}

	return resources, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
