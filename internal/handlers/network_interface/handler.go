package network_interface

import (
	"azsample/internal/az"
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

	client := clientFactory.NewInterfacesClient()

	nic, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	target := getAttachedResource(nic.Properties)

	if err != nil {
		return nil, err
	}

	properties := map[string]string{}
	dependsOn := []string{}

	if target != nil {
		t := strings.ToLower(*target)
		properties["attachedTo"] = t
		dependsOn = append(dependsOn, t)
	}

	subnet := nic.Properties.IPConfigurations[0].Properties.Subnet.ID

	if subnet != nil {
		s := strings.ToLower(*subnet)
		dependsOn = append(dependsOn, s)
	}

	resource := &az.Resource{
		Id:            *nic.ID,
		Name:          *nic.Name,
		Type:          *nic.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     dependsOn,
		Properties:    properties,
	}

	return []*az.Resource{resource}, nil
}

func getAttachedResource(nic *armnetwork.InterfacePropertiesFormat) *string {
	if nic.PrivateEndpoint != nil {
		return nic.PrivateEndpoint.ID
	}

	if nic.PrivateLinkService != nil {
		return nic.PrivateLinkService.ID
	}

	if nic.VirtualMachine != nil {
		return nic.VirtualMachine.ID
	}

	return nil
}
