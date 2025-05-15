package api_management_service

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement/v3"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armapimanagement.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewServiceClient()

	apim, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	dependsOn = append(dependsOn, strings.ToLower(*apim.Properties.PublicIPAddressID))
	dependsOn = append(dependsOn, strings.ToLower(*apim.Properties.VirtualNetworkConfiguration.SubnetResourceID))

	resources := []*models.Resource{&models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *apim.Type,
		DependsOn: dependsOn,
	}}

	apis, err := getAPIs(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	resources = append(resources, apis...)

	return resources, nil
}

func getAPIs(clientFactory *armapimanagement.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewAPIClient()

	pager := client.NewListByServicePager(ctx.ResourceGroupName, ctx.ResourceName, nil)

	var apis []*armapimanagement.APIContract
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.APICollection.Value != nil {
			apis = append(apis, resp.APICollection.Value...)
		}
	}

	return list.Map(apis, func(api *armapimanagement.APIContract) *models.Resource {
		dependsOn := []string{ctx.ResourceId}

		return &models.Resource{
			Id:        *api.ID,
			Name:      *api.Name,
			Type:      *api.Type,
			DependsOn: dependsOn,
		}
	}), nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
