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

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armnetwork.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewInterfacesClient()

	nic, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

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

	resource := &models.Resource{
		Id:         *nic.ID,
		Name:       *nic.Name,
		Type:       *nic.Type,
		DependsOn:  dependsOn,
		Properties: properties,
	}

	return []*models.Resource{resource}, nil
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
