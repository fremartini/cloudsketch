package context

import "github.com/Azure/azure-sdk-for-go/sdk/azidentity"

type Context struct {
	Credentials                                                           *azidentity.DefaultAzureCredential
	SubscriptionId, ResourceGroupName, ResourceName, ResourceId, TenantId string
}

type SubscriptionContext struct {
	Id, Name, TenantId string
}
