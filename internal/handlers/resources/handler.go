package resources

import (
	"azsample/internal/az"
	"azsample/internal/list"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) Handle(ctx *az.Context) ([]*az.Resource, error) {
	clientFactory, _ := armresources.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)
	client := clientFactory.NewClient()

	pager := client.NewListByResourceGroupPager(ctx.ResourceName, nil)

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

	azResources := list.Map(resources, func(resource *armresources.GenericResourceExpanded) *az.Resource {
		return &az.Resource{
			Id:            *resource.ID,
			Name:          *resource.Name,
			Type:          *resource.Type,
			ResourceGroup: ctx.ResourceGroup,
		}
	})

	return azResources, nil
}
