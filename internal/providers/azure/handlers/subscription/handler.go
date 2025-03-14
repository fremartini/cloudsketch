package subscription

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) Handle(subscriptionId string, credentials *azidentity.DefaultAzureCredential) (*azContext.SubscriptionContext, error) {
	clientFactory, err := armsubscriptions.NewClientFactory(credentials, nil)

	if err != nil {
		return nil, err
	}

	subscription, err := clientFactory.NewClient().Get(context.Background(), subscriptionId, nil)

	if err != nil {
		return nil, err
	}

	return &azContext.SubscriptionContext{
		Id:         *subscription.Subscription.SubscriptionID,
		ResourceId: *subscription.ID,
		Name:       *subscription.Subscription.DisplayName,
		TenantId:   *subscription.Subscription.TenantID,
	}, nil
}
