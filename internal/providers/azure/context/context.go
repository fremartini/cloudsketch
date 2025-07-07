package context

import "github.com/Azure/azure-sdk-for-go/sdk/azidentity"

type Context struct {
	Credentials                             *azidentity.DefaultAzureCredential
	TenantId, ResourceGroup                 string
	ManagementGroup, Subscription, Resource *ResourceIdentifier
}

type ResourceIdentifier struct {
	Name, Id, ResourceId string
}
