package web_sites

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v2"
)

type handler struct{}

var (
	WEBSITES_KIND_MAP = map[string]string{
		"app":                                    types.APP_SERVICE,
		"app,linux":                              types.APP_SERVICE,
		"app,linux,container":                    types.APP_SERVICE,
		"hyperV":                                 types.APP_SERVICE,
		"app,container,windows":                  types.APP_SERVICE,
		"app,linux,kubernete":                    types.APP_SERVICE,
		"app,linux,container,kubernetes":         types.APP_SERVICE,
		"functionapp":                            types.FUNCTION_APP,
		"functionapp,linux":                      types.FUNCTION_APP,
		"functionapp,linux,container,kubernetes": types.FUNCTION_APP,
		"functionapp,linux,kubernetes":           types.FUNCTION_APP,
		"functionapp,workflowapp":                types.LOGIC_APP,
	}
)

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armappservice.NewWebAppsClient(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	app, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	config, err := client.GetConfiguration(context.Background(), ctx.ResourceGroup, ctx.Resource.Name, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	resourceDependenciesInTags, err := getResourceReferencesInTags(ctx)

	if err != nil {
		return nil, err
	}

	dependsOn = append(dependsOn, resourceDependenciesInTags...)

	if app.Identity != nil {
		for identity := range app.Identity.UserAssignedIdentities {
			t := strings.ToLower(identity)
			dependsOn = append(dependsOn, t)
		}
	}

	properties := map[string][]string{}

	configValues := config.Properties.AzureStorageAccounts[ctx.Resource.Name]

	if configValues != nil {
		properties["storageAccountName"] = []string{strings.ToLower(*configValues.AccountName)}
	}

	// Microsoft.Web/sites has multiple subcategories. Use these instead
	subType := WEBSITES_KIND_MAP[*app.Kind]

	outboundSubnetId := app.Properties.VirtualNetworkSubnetID

	if outboundSubnetId != nil {
		properties["outboundSubnet"] = []string{strings.ToLower(*outboundSubnetId)}
	}

	planId := app.Properties.ServerFarmID
	dependsOn = append(dependsOn, *planId)

	resource.Type = subType
	resource.Properties = properties
	resource.DependsOn = append(resource.DependsOn, dependsOn...)

	return []*models.Resource{}, nil
}

func getResourceReferencesInTags(ctx *azContext.Context) ([]string, error) {
	clientFactory, err := armresources.NewClientFactory(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewTagsClient()

	tags, err := client.GetAtScope(context.Background(), ctx.Resource.Id, nil)

	if err != nil {
		return nil, err
	}

	result := []string{}
	tagsToKeep := []string{
		"hidden-link: /app-insights-resource-id",
	}

	for k, v := range tags.Properties.Tags {
		if !list.Contains(tagsToKeep, func(t string) bool {
			return k == t
		}) {
			continue
		}

		result = append(result, *v)
	}

	return result, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
