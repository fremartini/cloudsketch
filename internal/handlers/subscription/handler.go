package subscription

import (
	"cloudsketch/internal/az"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) Handle(subscriptionId string, credentials *azidentity.DefaultAzureCredential) (*az.SubscriptionContext, error) {
	clientFactory, err := armsubscriptions.NewClientFactory(credentials, nil)

	if err != nil {
		return nil, err
	}

	client, err := clientFactory.NewClient().Get(context.Background(), subscriptionId, nil)

	if err != nil {
		return nil, err
	}

	return &az.SubscriptionContext{
		Id:   *client.Subscription.SubscriptionID,
		Name: *client.Subscription.DisplayName,
	}, nil
}
