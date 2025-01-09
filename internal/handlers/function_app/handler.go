package function_app

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

	fa, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	asp := fa.Properties.ServerFarmID

	resource := &az.Resource{
		Id:            *fa.ID,
		Name:          *fa.Name,
		Type:          *fa.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     []string{*asp},
	}

	return []*az.Resource{resource}, nil
}
