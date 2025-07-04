package subscription

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/handlers/resource_group"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	log.Printf("fetching resources in subscription %s", ctx.ResourceId)

	clientFactory, err := armsubscriptions.NewClientFactory(ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	subscription, err := clientFactory.NewClient().Get(context.Background(), ctx.ResourceId, nil)

	if err != nil {
		return nil, err
	}

	resourceGroups, err := getResourceGroupsInSubscription(ctx)

	if err != nil {
		return nil, err
	}

	resources := []*models.Resource{
		{
			Id:        *subscription.ID,
			Name:      *subscription.DisplayName,
			Type:      types.SUBSCRIPTION,
			DependsOn: []string{ctx.ManagementGroupId},
		},
	}

	for _, resourceGroup := range resourceGroups {
		resourceGroupResources, err := resource_group.New().Handle(&azContext.Context{
			Credentials:       ctx.Credentials,
			ResourceName:      resourceGroup.Name,
			SubscriptionId:    ctx.ResourceId,
			ManagementGroupId: ctx.ManagementGroupId,
		})

		if err != nil {
			return nil, err
		}

		resources = append(resources, resourceGroupResources...)
	}

	return resources, nil
}

func getResourceGroupsInSubscription(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armresources.NewResourceGroupsClient(ctx.ResourceId, ctx.Credentials, nil)

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

	resources := list.Map(resourceGroups, func(resourceGroup *armresources.ResourceGroup) *models.Resource {
		return &models.Resource{
			Id:            *resourceGroup.ID,
			Name:          *resourceGroup.Name,
			Type:          *resourceGroup.Type,
			ResourceGroup: *resourceGroup.Name,
			DependsOn:     []string{ctx.ResourceId},
		}
	})

	return resources, nil
}
