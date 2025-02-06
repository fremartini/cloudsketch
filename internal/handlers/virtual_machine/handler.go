package virtual_machine

import (
	"azsample/internal/az"
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

	client := clientFactory.NewVirtualMachinesClient()

	vm, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	for _, nic := range vm.Properties.NetworkProfile.NetworkInterfaces {
		t := strings.ToLower(*nic.ID)
		dependsOn = append(dependsOn, t)
	}

	resources := []*az.Resource{
		{
			Id:            *vm.ID,
			Name:          *vm.Name,
			Type:          *vm.Type,
			ResourceGroup: ctx.ResourceGroup,
			DependsOn:     dependsOn,
		},
	}

	return resources, nil
}
