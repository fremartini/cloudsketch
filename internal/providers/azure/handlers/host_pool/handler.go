package host_pool

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/desktopvirtualization/armdesktopvirtualization/v2"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armdesktopvirtualization.NewSessionHostsClient(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(ctx.ResourceGroup, ctx.Resource.Name, nil)

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

	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return []*models.Resource{}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
