package application_gateway

import (
	"cloudsketch/internal/az"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *az.Context) ([]*az.Resource, error) {
	client, err := armnetwork.NewApplicationGatewaysClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	agw, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	if subnet := getSubnet(&agw); subnet != nil {
		dependsOn = append(dependsOn, *subnet)
	}

	if publicIp := getPublicIpAddress(&agw); publicIp != nil {
		dependsOn = append(dependsOn, *publicIp)
	}

	resource := &az.Resource{
		Id:            *agw.ID,
		Name:          *agw.Name,
		Type:          *agw.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     dependsOn,
	}

	return []*az.Resource{resource}, nil
}

func getSubnet(agw *armnetwork.ApplicationGatewaysClientGetResponse) *string {
	return agw.Properties.GatewayIPConfigurations[0].Properties.Subnet.ID
}

func getPublicIpAddress(agw *armnetwork.ApplicationGatewaysClientGetResponse) *string {
	frontends := agw.ApplicationGateway.Properties.FrontendIPConfigurations

	for _, frontendIpConfig := range frontends {
		publicAddress := frontendIpConfig.Properties.PublicIPAddress

		if publicAddress == nil {
			continue
		}

		return publicAddress.ID
	}

	return nil
}
