package subscription

import (
	"cloudsketch/internal/list"
	"cloudsketch/internal/providers/azure/containers/resource_group"
	azContext "cloudsketch/internal/providers/azure/context"
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
	log.Printf("fetching resources in subscription %s", ctx.Resource.Id)

	clientFactory, err := armsubscriptions.NewClientFactory(ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	subscription, err := clientFactory.NewClient().Get(context.Background(), ctx.Resource.Id, nil)

	if err != nil {
		return nil, err
	}

	subscriptionResourceId := *subscription.Subscription.ID

	ctx.Resource.Name = *subscription.DisplayName
	ctx.Resource.ResourceId = subscriptionResourceId

	resourceGroups, err := getResourceGroupsInSubscription(ctx)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	if ctx.ManagementGroup != nil {
		dependsOn = append(dependsOn, ctx.ManagementGroup.Id)
	}

	resources := []*models.Resource{
		{
			Id:        *subscription.ID,
			Name:      *subscription.DisplayName,
			Type:      types.SUBSCRIPTION,
			DependsOn: dependsOn,
		},
	}

	for _, resourceGroup := range resourceGroups {
		rgCtx := &azContext.Context{
			Credentials: ctx.Credentials,
			Resource: &azContext.ResourceIdentifier{
				Id:         resourceGroup.Id,
				Name:       resourceGroup.Name,
				ResourceId: resourceGroup.Id,
			},
			Subscription: &azContext.ResourceIdentifier{
				Id:         *subscription.SubscriptionID,
				Name:       *subscription.DisplayName,
				ResourceId: subscriptionResourceId,
			},
			ManagementGroup: ctx.ManagementGroup,
		}

		resourceGroupResources, err := resource_group.New().Handle(rgCtx)

		if err != nil {
			return nil, err
		}

		resources = append(resources, resourceGroupResources...)
	}

	return resources, nil
}

func getResourceGroupsInSubscription(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armresources.NewResourceGroupsClient(ctx.Resource.Id, ctx.Credentials, nil)

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
			DependsOn:     []string{ctx.Resource.ResourceId},
		}
	})

	return resources, nil
}
