package management_group

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"log"
	"strings"

	"cloudsketch/internal/providers/azure/containers/subscription"

	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	log.Printf("fetching resources in management group %s", ctx.Resource.Id)

	clientFactory, err := armmanagementgroups.NewClientFactory(ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	managementGroup, err := clientFactory.NewClient().Get(context.Background(), ctx.Resource.Id, nil)

	if err != nil {
		return nil, err
	}

	ctx.Resource.Id = *managementGroup.Name
	ctx.Resource.ResourceId = *managementGroup.ID
	ctx.Resource.Name = *managementGroup.Properties.DisplayName

	itemsToProcess, err := getChildManagementGroupsAndSubscriptions(ctx, clientFactory)

	if err != nil {
		return nil, err
	}

	resources := []*models.Resource{}

	childManagementGroups, subscriptions := list.Split(itemsToProcess, func(resource *models.Resource) bool {
		return resource.Type == types.MANAGEMENT_GROUP
	})

	for _, childManagementGroup := range childManagementGroups {
		managementGroupResourceId := strings.Split(childManagementGroup.Id, "/")
		managementGroupId := managementGroupResourceId[len(managementGroupResourceId)-1]

		childManagementGroupCtx := &azContext.Context{
			Credentials: ctx.Credentials,
			Resource: &azContext.ResourceIdentifier{
				Id: managementGroupId,
			},
			ManagementGroup: ctx.Resource,
		}

		childResources, err := h.GetResource(childManagementGroupCtx)

		if err != nil {
			return nil, err
		}

		resources = append(resources, childResources...)
	}

	for _, childSubscription := range subscriptions {
		childSubscriptionResourceId := strings.Split(childSubscription.Id, "/")
		childSubscriptionId := childSubscriptionResourceId[len(childSubscriptionResourceId)-1]

		childSubscriptionCtx := &azContext.Context{
			Credentials: ctx.Credentials,
			Resource: &azContext.ResourceIdentifier{
				Id:         childSubscriptionId,
				Name:       childSubscription.Name,
				ResourceId: childSubscription.Id,
			},
			ManagementGroup: ctx.Resource,
		}

		childResources, err := subscription.New().GetResource(childSubscriptionCtx)

		if err != nil {
			return nil, err
		}

		resources = append(resources, childResources...)
	}

	dependsOn := []string{}

	if ctx.ManagementGroup != nil {
		dependsOn = append(dependsOn, ctx.ManagementGroup.ResourceId)
	}

	resources = append(resources, &models.Resource{
		Id:        *managementGroup.ID,
		Name:      *managementGroup.Properties.DisplayName,
		Type:      types.MANAGEMENT_GROUP,
		DependsOn: dependsOn,
	})

	return resources, nil
}

func getChildManagementGroupsAndSubscriptions(ctx *azContext.Context, clientFactory *armmanagementgroups.ClientFactory) ([]*models.Resource, error) {
	pager := clientFactory.NewClient().NewGetDescendantsPager(ctx.Resource.Id, nil)

	var descendants []*armmanagementgroups.DescendantInfo
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.DescendantListResult.Value != nil {
			descendants = append(descendants, resp.DescendantListResult.Value...)
		}
	}

	resources := list.Map(descendants, func(descendant *armmanagementgroups.DescendantInfo) *models.Resource {
		return &models.Resource{
			Id:        *descendant.ID,
			Name:      *descendant.Properties.DisplayName,
			Type:      *descendant.Type,
			DependsOn: []string{ctx.Resource.ResourceId},
		}
	})

	return resources, nil
}
