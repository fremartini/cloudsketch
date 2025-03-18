package application_gateway

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armnetwork.NewApplicationGatewaysClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	agw, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

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

	resource := &models.Resource{
		Id:        *agw.ID,
		Name:      *agw.Name,
		Type:      *agw.Type,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
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

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
