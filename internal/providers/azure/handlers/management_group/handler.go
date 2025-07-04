package management_group

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"log"
	"strings"

	"cloudsketch/internal/providers/azure/handlers/subscription"

	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	log.Printf("fetching resources in management group %s", ctx.ResourceId)

	clientFactory, err := armmanagementgroups.NewClientFactory(ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	managementGroup, err := clientFactory.NewClient().Get(context.Background(), ctx.ResourceId, nil)

	if err != nil {
		return nil, err
	}

	resources, err := getChildManagementGroupsAndSubscriptions(ctx.ResourceId, clientFactory)

	if err != nil {
		return nil, err
	}

	childManagementGroups, subscriptions := list.Split(resources, func(resource *models.Resource) bool {
		return resource.Type == types.MANAGEMENT_GROUP
	})

	for _, childManagementGroup := range childManagementGroups {
		managementGroupResourceId := strings.Split(childManagementGroup.Id, "/")
		managementGroupId := managementGroupResourceId[len(managementGroupResourceId)-1]

		childResources, err := h.GetResource(&azContext.Context{
			Credentials:       ctx.Credentials,
			ResourceId:        managementGroupId,
			ManagementGroupId: *managementGroup.ID,
		})

		if err != nil {
			return nil, err
		}

		resources = append(resources, childResources...)
	}

	for _, childSubscription := range subscriptions {
		childSubscriptionResourceId := strings.Split(childSubscription.Id, "/")
		childSubscriptionId := childSubscriptionResourceId[len(childSubscriptionResourceId)-1]

		childResources, err := subscription.New().GetResource(&azContext.Context{
			Credentials:       ctx.Credentials,
			ResourceId:        childSubscriptionId,
			ManagementGroupId: *managementGroup.ID,
		})

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
