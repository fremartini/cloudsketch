package azure

import (
	"cloudsketch/internal/concurrency"
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/list"
	"cloudsketch/internal/marshall"
	"cloudsketch/internal/providers"
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
	"cloudsketch/internal/providers/azure/handlers/resource_group"
	"cloudsketch/internal/providers/azure/handlers/subscription"
	"cloudsketch/internal/providers/azure/handlers/virtual_hub"
	"cloudsketch/internal/providers/azure/handlers/virtual_machine"
	"cloudsketch/internal/providers/azure/handlers/virtual_machine_scale_set"
	"cloudsketch/internal/providers/azure/handlers/virtual_network"
	"cloudsketch/internal/providers/azure/handlers/virtual_network_gateway"
	"cloudsketch/internal/providers/azure/handlers/web_sites"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"fmt"
	"log"
	"strings"

	domainTypes "cloudsketch/internal/frontends/types"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type handler interface {
	GetResource(ctx *azContext.Context) ([]*models.Resource, error)
	PostProcess(*models.Resource, []*models.Resource)
}

var (
	handlers map[string]handler = map[string]handler{
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
		types.VIRTUAL_HUB:                virtual_hub.New(),
		types.VIRTUAL_MACHINE:            virtual_machine.New(),
		types.VIRTUAL_MACHINE_SCALE_SET:  virtual_machine_scale_set.New(),
		types.VIRTUAL_NETWORK:            virtual_network.New(),
		types.VIRTUAL_NETWORK_GATEWAY:    virtual_network_gateway.New(),
		types.WEB_SITES:                  web_sites.New(),
	}
)

type azureProvider struct{}

func NewProvider() *azureProvider {
	return &azureProvider{}
}

func (h *azureProvider) FetchResources(subscriptionId string) ([]*providers.Resource, string, error) {
	credentials, err := azidentity.NewDefaultAzureCredential(nil)

	if err != nil {
		return nil, "", fmt.Errorf("authentication failure: %+v", err)
	}

	subscription, err := subscription.New().Handle(subscriptionId, credentials)

	if err != nil {
		return nil, "", err
	}

	ctx := &azContext.Context{
		SubscriptionId: subscription.Id,
		Credentials:    credentials,
		TenantId:       subscription.TenantId,
	}

	filename := fmt.Sprintf("%s_%s", subscription.Name, subscription.Id)
	filenameWithSuffix := fmt.Sprintf("%s.json", filename)

	cachedResources, ok := marshall.UnmarshalIfExists[[]*models.Resource](filenameWithSuffix)

	if ok {
		log.Printf("using existing file %s\n", filenameWithSuffix)

		return mapToProviderModel(*cachedResources), filename, nil
	}

	resources, err := fetchResources(subscription, ctx)

	if err != nil {
		return nil, "", err
	}

	postProcess(resources)

	addDependencyToSubscriptions(resources, subscription)

	resources = normalize(resources, ctx.TenantId, set.New[string]())

	// input resources can contain references to resources that do not exist (in other subscriptions for example). These need to be removed
	resources = filterUnknownDependencies(resources)

	return mapToProviderModel(resources), filename, nil
}

func mapToProviderModel(resources []*models.Resource) []*providers.Resource {
	return list.Map(resources, func(m *models.Resource) *providers.Resource {
		return &providers.Resource{
			Id:         m.Id,
			Name:       m.Name,
			Type:       m.Type,
			DependsOn:  m.DependsOn,
			Properties: m.Properties,
		}
	})
}

func normalize(resources []*models.Resource, tenantId string, unhandled_types *set.Set[string]) []*models.Resource {
	return list.Map(resources, func(resource *models.Resource) *models.Resource {
		return &models.Resource{
			Id:         strings.ToLower(resource.Id), // Azure is not consistent regarding casing. Ensure all id's are lowercase
			Type:       mapTypeToDomainType(resource.Type, unhandled_types),
			Name:       resource.Name,
			DependsOn:  list.Map(resource.DependsOn, strings.ToLower),
			Properties: linkOrDefault(resource, tenantId),
		}
	})
}

func linkOrDefault(resource *models.Resource, tenantId string) map[string][]string {
	properties := resource.Properties

	if properties == nil {
		properties = map[string][]string{}
	}

	link := generateAzurePortalLink(resource, tenantId)
	properties["link"] = []string{link}

	return properties
}

func addDependencyToSubscriptions(resources []*models.Resource, subscription *azContext.SubscriptionContext) {
	// all resources should have a dependency on the subscription. Except the subscription itself
	for _, resource := range resources {
		if resource.Id == subscription.ResourceId {
			continue
		}

		resource.DependsOn = append(resource.DependsOn, subscription.ResourceId)
	}
}

func fetchResources(subscription *azContext.SubscriptionContext, ctx *azContext.Context) ([]*models.Resource, error) {
	resources, err := resource_group.New().Handle(ctx)

	if err != nil {
		return nil, err
	}

	resourcesWithHandlers, resourcesWithoutHandlers := list.Split(resources, func(resource *models.Resource) bool {
		_, ok := handlers[resource.Type]

		return ok
	})

	functionsToApply := list.Map(resourcesWithHandlers, func(resource *models.Resource) func() ([]*models.Resource, error) {
		return func() ([]*models.Resource, error) {
			log.Print(resource.Name)

			handler := handlers[resource.Type]

			return handler.GetResource(&azContext.Context{
				SubscriptionId:    ctx.SubscriptionId,
				TenantId:          ctx.TenantId,
				Credentials:       ctx.Credentials,
				ResourceGroupName: resource.ResourceGroup,
				ResourceName:      resource.Name,
				ResourceId:        resource.Id,
			})
		}
	})

	resources, err = concurrency.FanOut(functionsToApply)

	if err != nil {
		return nil, err
	}

	// add the resources that don't have any handlers as-is
	resources = append(resources, resourcesWithoutHandlers...)

	// add the subscription entry
	resources = append(resources, &models.Resource{
		Id:   subscription.ResourceId,
		Name: subscription.Name,
		Type: types.SUBSCRIPTION,
	})

	return resources, nil
}

func postProcess(resources []*models.Resource) {
	for _, resource := range resources {
		handler, ok := handlers[resource.Type]

		if !ok {
			continue
		}

		handler.PostProcess(resource, resources)
	}
}

func generateAzurePortalLink(resource *models.Resource, tenant string) string {
	// https://portal.azure.com/#@<tenant>/resource/<resource id>
	return fmt.Sprintf("https://portal.azure.com/#@%s/resource%s", tenant, resource.Id)
}

func mapTypeToDomainType(azType string, unhandled_types *set.Set[string]) string {
	domainTypes := map[string]string{
		types.AI_SERVICES:                           domainTypes.AI_SERVICES,
		types.API_MANAGEMENT_API:                    domainTypes.API_MANAGEMENT_API,
		types.API_MANAGEMENT_SERVICE:                domainTypes.API_MANAGEMENT_SERVICE,
		types.APP_CONFIGURATION:                     domainTypes.APP_CONFIGURATION,
		types.APP_SERVICE:                           domainTypes.APP_SERVICE,
		types.APP_SERVICE_PLAN:                      domainTypes.APP_SERVICE_PLAN,
		types.APPLICATION_GATEWAY:                   domainTypes.APPLICATION_GATEWAY,
		types.APPLICATION_GROUP:                     domainTypes.APPLICATION_GROUP,
		types.APPLICATION_INSIGHTS:                  domainTypes.APPLICATION_INSIGHTS,
		types.APPLICATION_SECURITY_GROUP:            domainTypes.APPLICATION_SECURITY_GROUP,
		types.BACKEND_ADDRESS_POOL:                  domainTypes.BACKEND_ADDRESS_POOL,
		types.BASTION:                               domainTypes.BASTION,
		types.CONTAINER_APP:                         domainTypes.CONTAINER_APP,
		types.CONTAINER_APPS_ENVIRONMENT:            domainTypes.CONTAINER_APPS_ENVIRONMENT,
		types.CONNECTION:                            domainTypes.CONNECTION,
		types.CONTAINER_REGISTRY:                    domainTypes.CONTAINER_REGISTRY,
		types.COSMOS:                                domainTypes.COSMOS,
		types.DATA_FACTORY:                          domainTypes.DATA_FACTORY,
		types.DATA_FACTORY_INTEGRATION_RUNTIME:      domainTypes.DATA_FACTORY_INTEGRATION_RUNTIME,
		types.DATA_FACTORY_MANAGED_PRIVATE_ENDPOINT: domainTypes.DATA_FACTORY_MANAGED_PRIVATE_ENDPOINT,
		types.DATABRICKS_WORKSPACE:                  domainTypes.DATABRICKS_WORKSPACE,
		types.DNS_RECORD:                            domainTypes.DNS_RECORD,
		types.PRIVATE_DNS_RESOLVER:                  domainTypes.PRIVATE_DNS_RESOLVER,
		types.EXPRESS_ROUTE_CIRCUIT:                 domainTypes.EXPRESS_ROUTE_CIRCUIT,
		types.EXPRESS_ROUTE_GATEWAY:                 domainTypes.EXPRESS_ROUTE_GATEWAY,
		types.FUNCTION_APP:                          domainTypes.FUNCTION_APP,
		types.HOST_POOL:                             domainTypes.HOST_POOL,
		types.KEY_VAULT:                             domainTypes.KEY_VAULT,
		types.LOAD_BALANCER:                         domainTypes.LOAD_BALANCER,
		types.LOAD_BALANCER_FRONTEND:                domainTypes.LOAD_BALANCER_FRONTEND,
		types.LOG_ANALYTICS:                         domainTypes.LOG_ANALYTICS,
		types.LOGIC_APP:                             domainTypes.LOGIC_APP,
		types.MACHINE_LEARNING_WORKSPACE:            domainTypes.MACHINE_LEARNING_WORKSPACE,
		types.NAT_GATEWAY:                           domainTypes.NAT_GATEWAY,
		types.NETWORK_INTERFACE:                     domainTypes.NETWORK_INTERFACE,
		types.NETWORK_SECURITY_GROUP:                domainTypes.NETWORK_SECURITY_GROUP,
		types.POSTGRES_FLEXIBLE_SERVER:              domainTypes.POSTGRES_SQL_SERVER,
		types.POSTGRES_SQL_SERVER:                   domainTypes.POSTGRES_SQL_SERVER,
		types.PRIVATE_DNS_ZONE:                      domainTypes.PRIVATE_DNS_ZONE,
		types.PRIVATE_ENDPOINT:                      domainTypes.PRIVATE_ENDPOINT,
		types.PRIVATE_LINK_SERVICE:                  domainTypes.PRIVATE_LINK_SERVICE,
		types.PUBLIC_IP_ADDRESS:                     domainTypes.PUBLIC_IP_ADDRESS,
		types.RECOVERY_SERVICE_VAULT:                domainTypes.RECOVERY_SERVICE_VAULT,
		types.REDIS:                                 domainTypes.REDIS,
		types.ROUTE_TABLE:                           domainTypes.ROUTE_TABLE,
		types.SEARCH_SERVICE:                        domainTypes.SEARCH_SERVICE,
		types.SIGNALR:                               domainTypes.SIGNALR,
		types.SQL_DATABASE:                          domainTypes.SQL_DATABASE,
		types.SQL_SERVER:                            domainTypes.SQL_SERVER,
		types.STATIC_WEB_APP:                        domainTypes.STATIC_WEB_APP,
		types.STORAGE_ACCOUNT:                       domainTypes.STORAGE_ACCOUNT,
		types.SUBNET:                                domainTypes.SUBNET,
		types.SUBSCRIPTION:                          domainTypes.SUBSCRIPTION,
		types.USER_ASSIGNED_IDENTITY:                domainTypes.USER_ASSIGNED_IDENTITY,
		types.VIRTUAL_HUB:                           domainTypes.VIRTUAL_HUB,
		types.VIRTUAL_MACHINE:                       domainTypes.VIRTUAL_MACHINE,
		types.VIRTUAL_MACHINE_SCALE_SET:             domainTypes.VIRTUAL_MACHINE_SCALE_SET,
		types.VIRTUAL_MACHINE_SCALE_SET_INSTANCE:    domainTypes.VIRTUAL_MACHINE_SCALE_SET_INSTANCE,
		types.VIRTUAL_NETWORK:                       domainTypes.VIRTUAL_NETWORK,
		types.VIRTUAL_NETWORK_GATEWAY:               domainTypes.VIRTUAL_NETWORK_GATEWAY,
		types.VIRTUAL_WAN:                           domainTypes.VIRTUAL_WAN,
		types.WORKSPACE:                             domainTypes.WORKSPACE,
	}

	domainType, ok := domainTypes[azType]

	if !ok {
		seenResourceType := unhandled_types.Contains(azType)

		// mechanism to prevent spamming the output with the same type
		if !seenResourceType {
			log.Printf("undefined mapping from Azure types %s to domain type", azType)
			unhandled_types.Add(azType)
		}

		return azType
	}

	return domainType
}

func filterUnknownDependencies(resources []*models.Resource) []*models.Resource {
	for _, resource := range resources {
		resource.DependsOn = list.Filter(resource.DependsOn, func(d string) bool {
			dependency := list.FirstOrDefault(resources, nil, func(r *models.Resource) bool {
				return r.Id == d
			})

			if dependency == nil {
				log.Printf("removed unknown resource %s\n", d)
				return false
			}

			return true
		})
	}

	return resources
}
