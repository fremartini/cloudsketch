package virtual_machine_scale_set

import (
	"cloudsketch/internal/list"
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

	subnet := strings.ToLower(*vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations[0].Properties.IPConfigurations[0].Properties.Subnet.ID)

	dependsOn = append(dependsOn, subnet)

	resources := []*models.Resource{
		{
			Id:        *vmss.ID,
			Name:      *vmss.Name,
			Type:      *vmss.Type,
			DependsOn: dependsOn,
		},
	}

	vmssInstances, err := getInstances(ctx, clientFactory)

	if err != nil {
		return nil, err
	}

	resources = append(resources, vmssInstances...)

	return resources, nil
}

func getInstances(ctx *azContext.Context, clientFactory *armcompute.ClientFactory) ([]*models.Resource, error) {
	vmClient := clientFactory.NewVirtualMachineScaleSetVMsClient()

	pager := vmClient.NewListPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

	var instances []*armcompute.VirtualMachineScaleSetVM
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.VirtualMachineScaleSetVMListResult.Value != nil {
			instances = append(instances, resp.VirtualMachineScaleSetVMListResult.Value...)
		}
	}

	return list.Map(instances, func(instance *armcompute.VirtualMachineScaleSetVM) *models.Resource {
		subnet := instance.Properties.NetworkProfileConfiguration.NetworkInterfaceConfigurations[0].Properties.IPConfigurations[0].Properties.Subnet.ID

		dependsOn := []string{ctx.ResourceId, *subnet}

		return &models.Resource{
			Id:        *instance.ID,
			Name:      *instance.Name,
			Type:      *instance.Type,
			DependsOn: dependsOn,
		}
	}), nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
