package management_group

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) Handle(managementGroupId string, credentials *azidentity.DefaultAzureCredential) (*azContext.SubscriptionContext, error) {
	clientFactory, err := armmanagementgroups.NewClientFactory(credentials, nil)

	if err != nil {
		return nil, err
	}

	managementGroup, err := clientFactory.NewClient().Get(context.Background(), managementGroupId, nil)

	if err != nil {
		return nil, err
	}

	fmt.Println(managementGroup)

	childManagementGroupsAndSubscriptions, err := getDescendants(managementGroupId, clientFactory)

	if err != nil {
		return nil, err
	}

	fmt.Println(childManagementGroupsAndSubscriptions)

	return &azContext.SubscriptionContext{}, nil
}

func getDescendants(managementGroupId string, clientFactory *armmanagementgroups.ClientFactory) ([]*armmanagementgroups.DescendantInfo, error) {
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

	return descendants, nil
}
