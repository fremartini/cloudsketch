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

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armappservice.NewWebAppsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	app, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	config, err := client.GetConfiguration(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	dependsOn := []string{}

	resourceDependenciesInTags, err := getResourceReferencesInTags(ctx)

	if err != nil {
		return nil, err
	}

	dependsOn = append(dependsOn, resourceDependenciesInTags...)

	for identity := range app.Identity.UserAssignedIdentities {
		t := strings.ToLower(identity)
		dependsOn = append(dependsOn, t)
	}

	properties := map[string]any{}

	configValues := config.Properties.AzureStorageAccounts[ctx.ResourceName]

	if configValues != nil {
		properties["storageAccountName"] = strings.ToLower(*configValues.AccountName)
	}

	// Microsoft.Web/sites has multiple subcategories. Use these instead
	subType := WEBSITES_KIND_MAP[*app.Kind]

	outboundSubnetId := app.Properties.VirtualNetworkSubnetID
	properties["outboundSubnet"] = strings.ToLower(*outboundSubnetId)

	planId := app.Properties.ServerFarmID
	dependsOn = append(dependsOn, *planId)

	resource := &models.Resource{
		Id:         *app.ID,
		Name:       *app.Name,
		Type:       subType,
		DependsOn:  dependsOn,
		Properties: properties,
	}

	return []*models.Resource{resource}, nil
}

func getResourceReferencesInTags(ctx *azContext.Context) ([]string, error) {
	clientFactory, err := armresources.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewTagsClient()

	tags, err := client.GetAtScope(context.Background(), ctx.ResourceId, nil)

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
