package subscription

import (
	"azsample/internal/az"
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	ctx = context.Background()
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) Handle(subscriptionId string, credentials *azidentity.DefaultAzureCredential) az.SubscriptionContext {
	client, err := armresources.NewClient(subscriptionId, credentials, nil)

	if err != nil {
		log.Fatalf("Failed to create Subscription client: %v", err)
	}

	pager := client.NewListPager()
	for pager.More() {
		page, _ := pager.NextPage(context.Background())

		for _, subscription := range page.Value {
			return az.SubscriptionContext{
				Id:   *subscription.ID,
				Name: *subscription.Name,
			}
		}
	}
}

// func (*handler) Handle(subscriptionId string, credentials *azidentity.DefaultAzureCredential) ([]*armresources.ResourceGroup, error) {
// 	client, _ := armresources.NewResourceGroupsClient(subscriptionId, credentials, nil)
// 	pager := client.NewListPager(nil)

// 	var resourceGroups []*armresources.ResourceGroup
// 	for pager.More() {
// 		resp, err := pager.NextPage(ctx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if resp.ResourceGroupListResult.Value != nil {
// 			resourceGroups = append(resourceGroups, resp.ResourceGroupListResult.Value...)
// 		}
// 	}

// 	return resourceGroups, nil
// }
