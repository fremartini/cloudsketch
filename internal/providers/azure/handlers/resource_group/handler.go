package resource_group

import (
	"context"

	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) Handle(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armresources.NewResourceGroupsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(nil)

	var resourceGroups []*armresources.ResourceGroup

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		if resp.ResourceGroupListResult.Value != nil {
			resourceGroups = append(resourceGroups, resp.ResourceGroupListResult.Value...)
		}
	}

	resourceClient, err := armresources.NewClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	resources := []*models.Resource{}

	for _, resourceGroup := range resourceGroups {
		r, err := GetResourcesInResourceGroup(resourceClient, *resourceGroup.Name)

		if err != nil {
			return nil, err
		}

		resources = append(resources, r...)
	}

	return resources, nil
}

func GetResourcesInResourceGroup(client *armresources.Client, resourceGroup string) ([]*models.Resource, error) {
	pager := client.NewListByResourceGroupPager(resourceGroup, nil)

	var resources []*armresources.GenericResourceExpanded
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.ResourceListResult.Value != nil {
			resources = append(resources, resp.ResourceListResult.Value...)
		}
	}

	azResources := list.Map(resources, func(resource *armresources.GenericResourceExpanded) *models.Resource {
		return &models.Resource{
			Id:            *resource.ID,
			Name:          *resource.Name,
			Type:          *resource.Type,
			ResourceGroup: resourceGroup,
		}
	})

	return azResources, nil
}
