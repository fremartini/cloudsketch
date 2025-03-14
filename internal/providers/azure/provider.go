package azure

import (
	"cloudsketch/internal/concurrent"
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/list"
	"cloudsketch/internal/marshall"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/handlers/application_gateway"
	"cloudsketch/internal/providers/azure/handlers/application_insights"
	"cloudsketch/internal/providers/azure/handlers/bastion"
	"cloudsketch/internal/providers/azure/handlers/data_factory"
	"cloudsketch/internal/providers/azure/handlers/key_vault"
	"cloudsketch/internal/providers/azure/handlers/load_balancer"
	"cloudsketch/internal/providers/azure/handlers/nat_gateway"
	"cloudsketch/internal/providers/azure/handlers/network_interface"
	"cloudsketch/internal/providers/azure/handlers/private_dns_zone"
	"cloudsketch/internal/providers/azure/handlers/private_endpoint"
	"cloudsketch/internal/providers/azure/handlers/private_link_service"
	"cloudsketch/internal/providers/azure/handlers/resource_group"
	"cloudsketch/internal/providers/azure/handlers/subscription"
	"cloudsketch/internal/providers/azure/handlers/virtual_machine"
	"cloudsketch/internal/providers/azure/handlers/virtual_machine_scale_set"
	"cloudsketch/internal/providers/azure/handlers/virtual_network"
	"cloudsketch/internal/providers/azure/handlers/web_sites"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"fmt"
	"log"
	"strings"

	domainModels "cloudsketch/internal/drawio/models"
	domainTypes "cloudsketch/internal/drawio/types"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type handler interface {
	Handle(ctx *azContext.Context) ([]*models.Resource, error)
}

var (
	handlers map[string]handler = map[string]handler{
		types.APPLICATION_GATEWAY:       application_gateway.New(),
		types.APPLICATION_INSIGHTS:      application_insights.New(),
		types.DATA_FACTORY:              data_factory.New(),
		types.BASTION:                   bastion.New(),
		types.KEY_VAULT:                 key_vault.New(),
		types.LOAD_BALANCER:             load_balancer.New(),
		types.NAT_GATEWAY:               nat_gateway.New(),
		types.NETWORK_INTERFACE:         network_interface.New(),
		types.PRIVATE_DNS_ZONE:          private_dns_zone.New(),
		types.PRIVATE_ENDPOINT:          private_endpoint.New(),
		types.PRIVATE_LINK_SERVICE:      private_link_service.New(),
		types.VIRTUAL_MACHINE:           virtual_machine.New(),
		types.VIRTUAL_MACHINE_SCALE_SET: virtual_machine_scale_set.New(),
		types.VIRTUAL_NETWORK:           virtual_network.New(),
		types.WEB_SITES:                 web_sites.New(),
	}
)

type azureProvider struct{}

func NewProvider() *azureProvider {
	return &azureProvider{}
}

func (h *azureProvider) FetchResources(subscriptionId string) ([]*domainModels.Resource, string, error) {
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

	cachedResources, ok := marshall.UnmarshalIfExists[[]*domainModels.Resource](filenameWithSuffix)

	if ok {
		log.Printf("using existing file %s\n", filenameWithSuffix)

		return *cachedResources, filename, nil
	}

	resources, err := fetchAndMapResources(subscription, ctx)

	if err != nil {
		return nil, "", err
	}

	// cache resources for next run
	err = marshall.MarshallResources(filenameWithSuffix, resources)

	if err != nil {
		return nil, "", err
	}

	return resources, filename, nil
}

func fetchAndMapResources(subscription *azContext.SubscriptionContext, ctx *azContext.Context) ([]*domainModels.Resource, error) {
	resources, err := resource_group.New().Handle(ctx)

	if err != nil {
		return nil, err
	}

	resourcesWithHandlers, resourcesWithoutHandlers := list.GroupBy(resources, func(resource *models.Resource) bool {
		_, ok := handlers[resource.Type]

		return ok
	})

	functionsToApply := list.Map(resourcesWithHandlers, func(resource *models.Resource) func() ([]*models.Resource, error) {
		return func() ([]*models.Resource, error) {
			log.Print(resource.Name)

			handler := handlers[resource.Type]

			return handler.Handle(&azContext.Context{
				SubscriptionId:    ctx.SubscriptionId,
				TenantId:          ctx.TenantId,
				Credentials:       ctx.Credentials,
				ResourceGroupName: resource.ResourceGroup,
				ResourceName:      resource.Name,
				ResourceId:        resource.Id,
			})
		}
	})

	resources, err = concurrent.FanOut(functionsToApply)

	if err != nil {
		return nil, err
	}

	// add the resources that don't have any handlers as-is
	resources = append(resources, resourcesWithoutHandlers...)

	// add the subscription entry
	resources = append(resources, &models.Resource{
		Id:   subscription.Id,
		Name: subscription.Name,
		Type: types.SUBSCRIPTION,
	})

	unhandled_types := set.New[string]()
	domainResources := list.Map(resources, func(r *models.Resource) *domainModels.Resource {
		return mapToDomainResource(r, ctx.TenantId, unhandled_types)
	})

	return domainResources, nil
}

func mapToDomainResource(resource *models.Resource, tenantId string, unhandled_types *set.Set[string]) *domainModels.Resource {
	properties := resource.Properties

	if properties == nil {
		properties = map[string]string{}
	}

	properties["link"] = generateAzurePortalLink(resource, tenantId)

	// Azure is not consistent regarding casing. Ensure all id's are lowercase
	return &domainModels.Resource{
		Id:         strings.ToLower(resource.Id),
		Type:       mapTypeToDomainType(resource.Type, unhandled_types),
		Name:       resource.Name,
		DependsOn:  list.Map(resource.DependsOn, strings.ToLower),
		Properties: properties,
	}
}

func generateAzurePortalLink(resource *models.Resource, tenant string) string {
	// https://portal.azure.com/#@<tenant>/resource/<resource id>
	return fmt.Sprintf("https://portal.azure.com/#@%s/resource%s", tenant, resource.Id)
}

func mapTypeToDomainType(azType string, unhandled_types *set.Set[string]) string {
	domainTypes := map[string]string{
		types.AI_SERVICES:                           domainTypes.AI_SERVICES,
		types.APP_SERVICE:                           domainTypes.APP_SERVICE,
		types.APP_SERVICE_PLAN:                      domainTypes.APP_SERVICE_PLAN,
		types.APPLICATION_GATEWAY:                   domainTypes.APPLICATION_GATEWAY,
		types.APPLICATION_INSIGHTS:                  domainTypes.APPLICATION_INSIGHTS,
		types.APPLICATION_SECURITY_GROUP:            domainTypes.APPLICATION_SECURITY_GROUP,
		types.BASTION:                               domainTypes.BASTION,
		types.CONTAINER_REGISTRY:                    domainTypes.CONTAINER_REGISTRY,
		types.COSMOS:                                domainTypes.COSMOS,
		types.DATA_FACTORY:                          domainTypes.DATA_FACTORY,
		types.DATA_FACTORY_INTEGRATION_RUNTIME:      domainTypes.DATA_FACTORY_INTEGRATION_RUNTIME,
		types.DATA_FACTORY_MANAGED_PRIVATE_ENDPOINT: domainTypes.DATA_FACTORY_MANAGED_PRIVATE_ENDPOINT,
		types.DATABRICKS_WORKSPACE:                  domainTypes.DATABRICKS_WORKSPACE,
		types.DNS_RECORD:                            domainTypes.DNS_RECORD,
		types.FUNCTION_APP:                          domainTypes.FUNCTION_APP,
		types.KEY_VAULT:                             domainTypes.KEY_VAULT,
		types.LOAD_BALANCER:                         domainTypes.LOAD_BALANCER,
		types.LOAD_BALANCER_FRONTEND:                domainTypes.LOAD_BALANCER_FRONTEND,
		types.LOG_ANALYTICS:                         domainTypes.LOG_ANALYTICS,
		types.LOGIC_APP:                             domainTypes.LOGIC_APP,
		types.MACHINE_LEARNING_WORKSPACE:            domainTypes.MACHINE_LEARNING_WORKSPACE,
		types.NAT_GATEWAY:                           domainTypes.NAT_GATEWAY,
		types.NETWORK_INTERFACE:                     domainTypes.NETWORK_INTERFACE,
		types.NETWORK_SECURITY_GROUP:                domainTypes.NETWORK_SECURITY_GROUP,
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
		types.VIRTUAL_MACHINE:                       domainTypes.VIRTUAL_MACHINE,
		types.VIRTUAL_MACHINE_SCALE_SET:             domainTypes.VIRTUAL_MACHINE_SCALE_SET,
		types.VIRTUAL_NETWORK:                       domainTypes.VIRTUAL_NETWORK,
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
