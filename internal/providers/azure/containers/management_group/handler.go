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

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, string, error) {
	log.Printf("fetching resources in management group %s", ctx.Resource.Id)

	clientFactory, err := armmanagementgroups.NewClientFactory(ctx.Credentials, nil)

	if err != nil {
		return nil, "", err
	}

	managementGroup, err := clientFactory.NewClient().Get(context.Background(), ctx.Resource.Id, nil)

	if err != nil {
		return nil, "", err
	}

	ctx.Resource.Id = *managementGroup.Name
	ctx.Resource.ResourceId = *managementGroup.ID
	ctx.Resource.Name = *managementGroup.Properties.DisplayName

	childManagementGroups, subscriptions, err := getChildManagementGroupsAndSubscriptions(ctx, clientFactory)

	if err != nil {
		return nil, "", err
	}

	resources := []*models.Resource{
		{
			Id:   *managementGroup.ID,
			Name: *managementGroup.Properties.DisplayName,
			Type: types.MANAGEMENT_GROUP,
		},
	}

	resources = append(resources, childManagementGroups...)

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
			ManagementGroup: &azContext.ResourceIdentifier{
				ResourceId: childSubscription.Properties["parentId"][0],
			},
		}

		childResources, _, err := subscription.New().GetResource(childSubscriptionCtx)

		if err != nil {
			return nil, "", err
		}

		resources = append(resources, childResources...)
	}

	tenantId := *managementGroup.Properties.TenantID

	return resources, tenantId, nil
}

func getChildManagementGroupsAndSubscriptions(ctx *azContext.Context, clientFactory *armmanagementgroups.ClientFactory) ([]*models.Resource, []*models.Resource, error) {
	pager := clientFactory.NewClient().NewGetDescendantsPager(ctx.Resource.Id, nil)

	var descendants []*armmanagementgroups.DescendantInfo
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, nil, err
		}

		if resp.DescendantListResult.Value != nil {
			descendants = append(descendants, resp.DescendantListResult.Value...)
		}
	}

	resources := list.Map(descendants, func(descendant *armmanagementgroups.DescendantInfo) *models.Resource {
		resource := &models.Resource{
			Id:        *descendant.ID,
			Name:      *descendant.Properties.DisplayName,
			Type:      *descendant.Type,
			DependsOn: []string{*descendant.Properties.Parent.ID},
		}

		if resource.Type == types.SUBSCRIPTION {
			resource.Properties = map[string][]string{ // subscriptions are delegated to their appropriate handler. They need a reference to their parent
				"parentId": {*descendant.Properties.Parent.ID},
			}
		}

		return resource
	})

	managementGroups, subscriptions := list.Split(resources, func(resource *models.Resource) bool {
		return resource.Type == types.MANAGEMENT_GROUP
	})

	return managementGroups, subscriptions, nil
}
