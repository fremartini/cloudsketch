package sql_server

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql/v2"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(resource *models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armsql.NewClientFactory(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	databases, err := getDatabases(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	return databases, nil
}

func getDatabases(clientFactory *armsql.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewDatabasesClient()

	pager := client.NewListByServerPager(ctx.ResourceGroup, ctx.Resource.Name, nil)

	var databases []*armsql.Database
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.DatabaseListResult.Value != nil {
			databases = append(databases, resp.DatabaseListResult.Value...)
		}
	}

	dependsOn := []string{ctx.Resource.Id}

	models := list.Map(databases, func(database *armsql.Database) *models.Resource {
		return &models.Resource{
			Id:        *database.ID,
			Name:      *database.Name,
			Type:      *database.Type,
			DependsOn: dependsOn,
		}
	})

	return models, nil
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {

}
