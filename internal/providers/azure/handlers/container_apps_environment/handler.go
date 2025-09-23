package container_apps_environment

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"strings"

	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v3"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armappcontainers.NewManagedEnvironmentsClient(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	cae, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	if cae.Properties.VnetConfiguration != nil {
		subnet := cae.Properties.VnetConfiguration.InfrastructureSubnetID

		dependsOn = append(dependsOn, strings.ToLower(*subnet))
	}

	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return []*models.Resource{}, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
