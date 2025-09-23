package application_insights

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armapplicationinsights.NewComponentsClient(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	ai, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	if ai.Properties.WorkspaceResourceID != nil {
		dependsOn = append(dependsOn, *ai.Properties.WorkspaceResourceID)
	}

	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return []*models.Resource{}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
