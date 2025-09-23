package resource_group

import (
	"context"
	"log"

	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/handlers/api_management_service"
	"cloudsketch/internal/providers/azure/handlers/application_gateway"
	"cloudsketch/internal/providers/azure/handlers/application_group"
	"cloudsketch/internal/providers/azure/handlers/application_insights"
	"cloudsketch/internal/providers/azure/handlers/bastion"
	"cloudsketch/internal/providers/azure/handlers/container_app"
	"cloudsketch/internal/providers/azure/handlers/container_apps_environment"
	"cloudsketch/internal/providers/azure/handlers/data_factory"
	"cloudsketch/internal/providers/azure/handlers/express_route_circuit"
	"cloudsketch/internal/providers/azure/handlers/express_route_gateway"
	"cloudsketch/internal/providers/azure/handlers/host_pool"
	"cloudsketch/internal/providers/azure/handlers/key_vault"
	"cloudsketch/internal/providers/azure/handlers/load_balancer"
	"cloudsketch/internal/providers/azure/handlers/nat_gateway"
	"cloudsketch/internal/providers/azure/handlers/network_interface"
	"cloudsketch/internal/providers/azure/handlers/postgres_flexible_server"
	"cloudsketch/internal/providers/azure/handlers/private_dns_resolver"
	"cloudsketch/internal/providers/azure/handlers/private_dns_zone"
	"cloudsketch/internal/providers/azure/handlers/private_endpoint"
	"cloudsketch/internal/providers/azure/handlers/private_link_service"
	"cloudsketch/internal/providers/azure/handlers/sql_server"
	"cloudsketch/internal/providers/azure/handlers/virtual_hub"
	"cloudsketch/internal/providers/azure/handlers/virtual_machine"
	"cloudsketch/internal/providers/azure/handlers/virtual_machine_scale_set"
	"cloudsketch/internal/providers/azure/handlers/virtual_network"
	"cloudsketch/internal/providers/azure/handlers/virtual_network_gateway"
	"cloudsketch/internal/providers/azure/handlers/web_sites"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

type resourceHandler interface {
	GetResource(*models.Resource, *azContext.Context) ([]*models.Resource, error)
	PostProcess(*models.Resource, []*models.Resource)
}

var (
	handlers map[string]resourceHandler = map[string]resourceHandler{
		types.API_MANAGEMENT_SERVICE:     api_management_service.New(),
		types.APPLICATION_GATEWAY:        application_gateway.New(),
		types.APPLICATION_GROUP:          application_group.New(),
		types.APPLICATION_INSIGHTS:       application_insights.New(),
		types.DATA_FACTORY:               data_factory.New(),
		types.EXPRESS_ROUTE_CIRCUIT:      express_route_circuit.New(),
		types.EXPRESS_ROUTE_GATEWAY:      express_route_gateway.New(),
		types.HOST_POOL:                  host_pool.New(),
		types.BASTION:                    bastion.New(),
		types.CONTAINER_APP:              container_app.New(),
		types.CONTAINER_APPS_ENVIRONMENT: container_apps_environment.New(),
		types.KEY_VAULT:                  key_vault.New(),
		types.LOAD_BALANCER:              load_balancer.New(),
		types.NAT_GATEWAY:                nat_gateway.New(),
		types.NETWORK_INTERFACE:          network_interface.New(),
		types.POSTGRES_FLEXIBLE_SERVER:   postgres_flexible_server.New(),
		types.PRIVATE_DNS_RESOLVER:       private_dns_resolver.New(),
		types.PRIVATE_DNS_ZONE:           private_dns_zone.New(),
		types.PRIVATE_ENDPOINT:           private_endpoint.New(),
		types.PRIVATE_LINK_SERVICE:       private_link_service.New(),
		types.SQL_SERVER:                 sql_server.New(),
		types.VIRTUAL_HUB:                virtual_hub.New(),
		types.VIRTUAL_MACHINE:            virtual_machine.New(),
		types.VIRTUAL_MACHINE_SCALE_SET:  virtual_machine_scale_set.New(),
		types.VIRTUAL_NETWORK:            virtual_network.New(),
		types.VIRTUAL_NETWORK_GATEWAY:    virtual_network_gateway.New(),
		types.WEB_SITES:                  web_sites.New(),
	}
)

func (*handler) Handle(ctx *azContext.Context) ([]*models.Resource, error) {
	log.Printf("fetching resources in resource group %s", ctx.Resource.Name)

	resources, err := getResourcesInResourceGroup(ctx)

	if err != nil {
		return nil, err
	}

	resourcesWithHandlerResources, err := invokeResourceHandlers(resources, ctx)

	if err != nil {
		return nil, err
	}

	postProcessResources(resourcesWithHandlerResources)

	return resourcesWithHandlerResources, nil
}

func getResourcesInResourceGroup(ctx *azContext.Context) ([]*models.Resource, error) {
	client, err := armresources.NewClient(ctx.Subscription.Id, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	pager := client.NewListByResourceGroupPager(ctx.Resource.Name, nil)

	var resources []*armresources.GenericResourceExpanded
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.ResourceListResult.Value != nil {
			resources = append(resources, resp.ResourceListResult.Value...)
		}
	}

	models := list.Map(resources, func(resource *armresources.GenericResourceExpanded) *models.Resource {
		dependsOn := []string{ctx.Subscription.ResourceId}

		if ctx.ManagementGroup != nil {
			dependsOn = append(dependsOn, ctx.ManagementGroup.ResourceId)
		}

		return &models.Resource{
			Id:            *resource.ID,
			Name:          *resource.Name,
			Type:          *resource.Type,
			ResourceGroup: ctx.Resource.Name,
			DependsOn:     dependsOn,
		}
	})

	return models, nil
}

func invokeResourceHandlers(resources []*models.Resource, ctx *azContext.Context) ([]*models.Resource, error) {
	resourcesWithHandlers, resourcesWithoutHandlers := list.Split(resources, func(resource *models.Resource) bool {
		_, ok := handlers[resource.Type]

		return ok
	})

	for _, resource := range resourcesWithHandlers {
		log.Print(resource.Name)

		handler := handlers[resource.Type]

		resourceCtx := &azContext.Context{
			Subscription:  ctx.Subscription,
			TenantId:      ctx.TenantId,
			Credentials:   ctx.Credentials,
			ResourceGroup: resource.ResourceGroup,
			Resource: &azContext.ResourceIdentifier{
				Id:   resource.Id,
				Name: resource.Name,
			},
		}

		subresources, err := handler.GetResource(resource, resourceCtx)

		if err != nil {
			return nil, err
		}

		resourcesWithoutHandlers = append(resourcesWithoutHandlers, resource)
		resourcesWithoutHandlers = append(resourcesWithoutHandlers, subresources...)
	}

	return resourcesWithoutHandlers, nil
}

func postProcessResources(resources []*models.Resource) {
	for _, resource := range resources {
		lookupType := resource.Type

		// WEB_SITES changes type into one of its sub-types
		subTypes := []string{types.APP_SERVICE, types.FUNCTION_APP, types.LOGIC_APP}

		if list.Contains(subTypes, func(typ string) bool {
			return resource.Type == typ
		}) {
			lookupType = types.WEB_SITES
		}

		handler, ok := handlers[lookupType]

		if !ok {
			continue
		}

		handler.PostProcess(resource, resources)
	}
}
