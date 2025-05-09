package host_pool

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/desktopvirtualization/armdesktopvirtualization/v2"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armdesktopvirtualization.NewSessionHostsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

	var sessionHosts []*armdesktopvirtualization.SessionHost

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		if resp.SessionHostList.Value != nil {
			sessionHosts = append(sessionHosts, resp.SessionHostList.Value...)
		}
	}

	dependsOn := []string{}

	for _, host := range sessionHosts {
		dependsOn = append(dependsOn, *host.Properties.ResourceID)
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      types.HOST_POOL,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
