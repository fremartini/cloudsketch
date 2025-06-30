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

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armsql.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewServersClient()

	sqlServer, err := client.Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	databases, err := getDatabases(clientFactory, ctx)

	if err != nil {
		return nil, err
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      *sqlServer.Type,
		DependsOn: []string{},
	}

	resources := []*models.Resource{resource}

	resources = append(resources, databases...)

	return resources, nil
}

func getDatabases(clientFactory *armsql.ClientFactory, ctx *azContext.Context) ([]*models.Resource, error) {
	client := clientFactory.NewDatabasesClient()

	pager := client.NewListByServerPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

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

	dependsOn := []string{ctx.ResourceId}

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
