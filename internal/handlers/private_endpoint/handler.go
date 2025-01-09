package private_endpoint

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

	client := clientFactory.NewPrivateEndpointsClient()

	pe, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	pe_target := pe.Properties.PrivateLinkServiceConnections[0].Properties.PrivateLinkServiceID
	pe_subnet := pe.Properties.Subnet.ID

	properties := map[string]string{}
	dependsOn := []string{}

	if pe_target != nil {
		t := strings.ToLower(*pe_target)
		properties["attachedTo"] = t
		dependsOn = append(dependsOn, t)
	}

	if pe_subnet != nil {
		t := strings.ToLower(*pe_subnet)
		dependsOn = append(dependsOn, t)
	}

	resource := &az.Resource{
		Id:            *pe.ID,
		Name:          *pe.Name,
		Type:          *pe.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     dependsOn,
		Properties:    properties,
	}

	return []*az.Resource{resource}, nil
}
