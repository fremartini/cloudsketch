package web_sites

import (
	"cloudsketch/internal/az"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/types"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v4"
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

func (h *handler) Handle(ctx *azContext.Context) ([]*az.Resource, error) {
	client, err := armappservice.NewWebAppsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	app, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	config, err := client.GetConfiguration(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	properties := map[string]string{}

	configValues := config.Properties.AzureStorageAccounts[ctx.ResourceName]

	if configValues != nil {
		properties["storageAccountName"] = strings.ToLower(*configValues.AccountName)
	}

	// Microsoft.Web/sites has multiple subcategories. Use these instead
	subType := WEBSITES_KIND_MAP[*app.Kind]

	outboundSubnetId := app.Properties.VirtualNetworkSubnetID
	properties["outboundSubnet"] = strings.ToLower(*outboundSubnetId)

	planId := app.Properties.ServerFarmID
	dependsOn := []string{*planId}

	resource := &az.Resource{
		Id:            *app.ID,
		Name:          *app.Name,
		Type:          subType,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     dependsOn,
		Properties:    properties,
	}

	return []*az.Resource{resource}, nil
}
