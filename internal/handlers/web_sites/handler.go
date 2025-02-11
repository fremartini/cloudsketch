package web_sites

import (
	"azsample/internal/az"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v4"
)

type handler struct{}

var (
	WEBSITES_KIND_MAP = map[string]string{
		"app":                                    az.APP_SERVICE_SUBTYPE,
		"app,linux":                              az.APP_SERVICE_SUBTYPE,
		"app,linux,container":                    az.APP_SERVICE_SUBTYPE,
		"hyperV":                                 az.APP_SERVICE_SUBTYPE,
		"app,container,windows":                  az.APP_SERVICE_SUBTYPE,
		"app,linux,kubernete":                    az.APP_SERVICE_SUBTYPE,
		"app,linux,container,kubernetes":         az.APP_SERVICE_SUBTYPE,
		"functionapp":                            az.FUNCTION_APP_SUBTYPE,
		"functionapp,linux":                      az.FUNCTION_APP_SUBTYPE,
		"functionapp,linux,container,kubernetes": az.FUNCTION_APP_SUBTYPE,
		"functionapp,linux,kubernetes":           az.FUNCTION_APP_SUBTYPE,
		"functionapp,workflowapp":                az.LOGIC_APP_SUBTYPE,
	}
)

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *az.Context) ([]*az.Resource, error) {
	client, err := armappservice.NewWebAppsClient(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	app, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	plan := app.Properties.ServerFarmID

	dependsOn := []string{*plan}

	subType := WEBSITES_KIND_MAP[*app.Kind]

	properties := map[string]string{"subType": subType}

	resource := &az.Resource{
		Id:            *app.ID,
		Name:          *app.Name,
		Type:          *app.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     dependsOn,
		Properties:    properties,
	}

	return []*az.Resource{resource}, nil
}
