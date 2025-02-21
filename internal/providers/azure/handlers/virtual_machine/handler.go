package virtual_machine

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

func (h *handler) Handle(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armcompute.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewVirtualMachinesClient()

	vm, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	for _, nic := range vm.Properties.NetworkProfile.NetworkInterfaces {
		t := strings.ToLower(*nic.ID)
		dependsOn = append(dependsOn, t)
	}

	resources := []*models.Resource{
		{
			Id:        *vm.ID,
			Name:      *vm.Name,
			Type:      *vm.Type,
			DependsOn: dependsOn,
		},
	}

	return resources, nil
}
