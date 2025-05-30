package private_endpoint

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

	client := clientFactory.NewPrivateEndpointsClient()

	pe, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	properties := map[string][]string{}
	dependsOn := []string{}

	if len(pe.Properties.PrivateLinkServiceConnections) > 0 {
		pe_target := pe.Properties.PrivateLinkServiceConnections[0].Properties.PrivateLinkServiceID

		if pe_target != nil {
			t := strings.ToLower(*pe_target)
			properties["attachedTo"] = []string{t}
			dependsOn = append(dependsOn, t)
		}
	}

	for _, nic := range pe.Properties.NetworkInterfaces {
		t := strings.ToLower(*nic.ID)
		dependsOn = append(dependsOn, t)
	}

	for _, asg := range pe.Properties.ApplicationSecurityGroups {
		t := strings.ToLower(*asg.ID)
		dependsOn = append(dependsOn, t)
	}

	pe_subnet := pe.Properties.Subnet.ID

	if pe_subnet != nil {
		t := strings.ToLower(*pe_subnet)
		dependsOn = append(dependsOn, t)
	}

	resource := &models.Resource{
		Id:         *pe.ID,
		Name:       *pe.Name,
		Type:       *pe.Type,
		DependsOn:  dependsOn,
		Properties: properties,
	}

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
