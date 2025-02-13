package virtual_machine_scale_set

import (
	"cloudsketch/internal/az"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *az.Context) ([]*az.Resource, error) {
	clientFactory, err := armcompute.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewVirtualMachineScaleSetsClient()

	vmss, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	subnet := strings.ToLower(*vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations[0].Properties.IPConfigurations[0].Properties.Subnet.ID)

	dependsOn = append(dependsOn, subnet)

	resources := []*az.Resource{
		{
			Id:            *vmss.ID,
			Name:          *vmss.Name,
			Type:          *vmss.Type,
			ResourceGroup: ctx.ResourceGroup,
			DependsOn:     dependsOn,
		},
	}

	return resources, nil
}
