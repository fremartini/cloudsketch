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

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armapimanagement.NewClientFactory(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewServiceClient()

	apim, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{
		strings.ToLower(*apim.Properties.PublicIPAddressID),
		strings.ToLower(*apim.Properties.VirtualNetworkConfiguration.SubnetResourceID),
	}

	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return getAPIs(clientFactory, ctx)
}

func getAPIs(clientFactory *armapimanagement.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewAPIClient()

	pager := client.NewListByServicePager(ctx.ResourceGroup, ctx.Resource.Name, nil)

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
		dependsOn := []string{ctx.Resource.Id}

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
