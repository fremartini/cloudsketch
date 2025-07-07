package network_interface

import (
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

	client := clientFactory.NewInterfacesClient()

	nic, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	properties := map[string][]string{}

	properties["ip"] = []string{*nic.Properties.IPConfigurations[0].Properties.PrivateIPAddress}

	target := getAttachedResource(nic.Properties)

	if target != nil {
		t := strings.ToLower(*target)
		properties["attachedTo"] = []string{t}
	}

	dependsOn := []string{}

	subnet := nic.Properties.IPConfigurations[0].Properties.Subnet.ID

	if subnet != nil {
		s := strings.ToLower(*subnet)
		dependsOn = append(dependsOn, s)
	}

	resource.DependsOn = append(resource.DependsOn, dependsOn...)
	resource.Properties = properties

	return []*models.Resource{}, nil
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

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
