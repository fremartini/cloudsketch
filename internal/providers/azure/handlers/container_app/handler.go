package container_app

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v3"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armappcontainers.NewContainerAppsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	ca, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	for identity := range ca.Identity.UserAssignedIdentities {
		t := strings.ToLower(identity)
		dependsOn = append(dependsOn, t)
	}

	// container apps environment
	dependsOn = append(dependsOn, strings.ToLower(*ca.Properties.EnvironmentID))

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *ca.Type,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
