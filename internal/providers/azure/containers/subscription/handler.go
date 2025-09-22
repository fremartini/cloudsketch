package subscription

import (
	"cloudsketch/internal/concurrency"
	"cloudsketch/internal/list"
	"cloudsketch/internal/providers/azure/containers/resource_group"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"
	"errors"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) GetResource(ctx *azContext.Context) ([]*models.Resource, string, error) {
	log.Printf("fetching resources in subscription %s", ctx.Resource.Id)

	clientFactory, err := armsubscriptions.NewClientFactory(ctx.Credentials, nil)

	if err != nil {
		return nil, "", err
	}

	subscription, err := clientFactory.NewClient().Get(context.Background(), ctx.Resource.Id, nil)

	if err != nil {
		// if the subscription is not found, treat it as success. It might have been deleted during the program runtime
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if respErr.StatusCode == 404 {
				log.Printf("subscription %s was not found", ctx.Resource.Id)

				return []*models.Resource{}, "", nil
			}
		}

		return nil, "", err
	}

	subscriptionResourceId := *subscription.Subscription.ID

	ctx.Resource.Name = *subscription.DisplayName
	ctx.Resource.ResourceId = subscriptionResourceId

	resourceGroups, err := getResourceGroupsInSubscription(ctx)

	if err != nil {
		return nil, "", err
	}

	dependsOn := []string{}

	if ctx.ManagementGroup != nil {
		dependsOn = append(dependsOn, ctx.ManagementGroup.ResourceId)
	}

	functionsToApply := list.Map(resourceGroups, func(resourceGroup *models.Resource) func() ([]*models.Resource, error) {
		return func() ([]*models.Resource, error) {
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

			return resourceGroupResources, nil
		}
	})

	resources, err := concurrency.FanOut(functionsToApply)

	if err != nil {
		return nil, "", err
	}

	return append([]*models.Resource{
		{
			Id:        *subscription.ID,
			Name:      *subscription.DisplayName,
			Type:      types.SUBSCRIPTION,
			DependsOn: dependsOn,
		},
	}, resources...), *subscription.TenantID, nil
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
