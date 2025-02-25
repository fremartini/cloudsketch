package key_vault

import (
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *azContext.Context) ([]*models.Resource, error) {
	diagnosticsClient, err := armmonitor.NewDiagnosticSettingsClient(ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pager := diagnosticsClient.NewListPager(ctx.ResourceId, nil)

	var resources []*armmonitor.DiagnosticSettingsResource
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.DiagnosticSettingsResourceCollection.Value != nil {
			resources = append(resources, resp.DiagnosticSettingsResourceCollection.Value...)
		}
	}

	dependsOn := []string{}
	for _, diagnostic := range resources {
		if diagnostic.Properties.WorkspaceID != nil {
			dependsOn = append(dependsOn, *diagnostic.Properties.WorkspaceID)
		}

		if diagnostic.Properties.StorageAccountID != nil {
			dependsOn = append(dependsOn, *diagnostic.Properties.StorageAccountID)
		}
	}

	resource := &models.Resource{
		Id:        ctx.ResourceId,
		Name:      ctx.ResourceName,
		Type:      types.KEY_VAULT,
		DependsOn: dependsOn,
	}

	return []*models.Resource{resource}, nil
}
