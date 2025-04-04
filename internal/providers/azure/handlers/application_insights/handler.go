package application_insights

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armapplicationinsights.NewComponentsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	ai, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	properties := map[string]any{}
	dependsOn := []string{}

	if ai.Properties.WorkspaceResourceID != nil {
		dependsOn = append(dependsOn, *ai.Properties.WorkspaceResourceID)
	}

	resource := &models.Resource{
		Id:         *ai.ID,
		Name:       *ai.Name,
		Type:       types.APPLICATION_INSIGHTS, // type returns lowercase microsoft.insights/actiongroups
		DependsOn:  dependsOn,
		Properties: properties,
	}

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
