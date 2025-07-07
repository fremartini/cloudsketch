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

	resources, err := getChildManagementGroupsAndSubscriptions(ctx.Resource.Id, clientFactory)

	if err != nil {
		return nil, err
	}

	childManagementGroups, subscriptions := list.Split(resources, func(resource *models.Resource) bool {
		return resource.Type == types.MANAGEMENT_GROUP
	})

	for _, childManagementGroup := range childManagementGroups {
		managementGroupResourceId := strings.Split(childManagementGroup.Id, "/")
		managementGroupId := managementGroupResourceId[len(managementGroupResourceId)-1]

		childManagementGroupCtx := &azContext.Context{
			Credentials: ctx.Credentials,
			Resource: &azContext.ResourceIdentifier{
				Id:   managementGroupId,
				Name: childManagementGroup.Name,
			},
			ManagementGroup: &azContext.ResourceIdentifier{
				Id:   *managementGroup.ID,
				Name: childManagementGroup.Name,
			},
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
				Id:   childSubscriptionId,
				Name: childSubscription.Name,
			},
			ManagementGroup: &azContext.ResourceIdentifier{
				Id:   *managementGroup.ID,
				Name: *managementGroup.Name,
			},
		}

		childResources, err := subscription.New().GetResource(childSubscriptionCtx)

		if err != nil {
			return nil, err
		}

		resources = append(resources, childResources...)
	}

	resources = append(resources, &models.Resource{
		Id:        *managementGroup.ID,
		Name:      *managementGroup.Properties.DisplayName,
		Type:      types.MANAGEMENT_GROUP,
		DependsOn: []string{},
	})

	return resources, nil
}

func getChildManagementGroupsAndSubscriptions(managementGroupId string, clientFactory *armmanagementgroups.ClientFactory) ([]*models.Resource, error) {
	pager := clientFactory.NewClient().NewGetDescendantsPager(managementGroupId, nil)

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
			DependsOn: []string{managementGroupId},
		}
	})

	return resources, nil
}
