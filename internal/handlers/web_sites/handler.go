package web_sites

import (
	"azsample/internal/az"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v4"
)

type handler struct{}

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
	subType := app.Kind

	resource := &az.Resource{
		Id:            *app.ID,
		Name:          *app.Name,
		Type:          *app.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     []string{*plan},
		Properties:    map[string]string{"SubType": *subType},
	}

	return []*az.Resource{resource}, nil
}
