package az

import "github.com/Azure/azure-sdk-for-go/sdk/azidentity"

type Context struct {
	Credentials    *azidentity.DefaultAzureCredential
	SubscriptionId string
	ResourceGroup  string
	ResourceName   string
	ResourceId     string
}
