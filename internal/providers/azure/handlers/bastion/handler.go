package bastion

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

func (h *handler) Handle(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armnetwork.NewBastionHostsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	bastion, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	for _, config := range bastion.Properties.IPConfigurations {
		dependsOn = append(dependsOn, *config.Properties.PublicIPAddress.ID)
		dependsOn = append(dependsOn, *config.Properties.Subnet.ID)
		dependsOn = append(dependsOn, *config.Properties.PublicIPAddress.ID)
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *bastion.Type,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
}
