package subscription

import (
	"azsample/internal/az"
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (*handler) Handle(subscriptionId string, credentials *azidentity.DefaultAzureCredential) az.SubscriptionContext {
	clientFactory, err := armsubscriptions.NewClientFactory(credentials, nil)
	client, err := clientFactory.NewClient().Get(context.Background(), subscriptionId, nil)

	if err != nil {
		log.Fatalf("Failed to create Subscription client: %v", err)
	}

	return az.SubscriptionContext{
		Id:   *client.Subscription.SubscriptionID,
		Name: *client.Subscription.DisplayName,
	}
}
