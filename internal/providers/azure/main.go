package azure

import (
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/guid"
	"cloudsketch/internal/list"
	"cloudsketch/internal/marshall"
	"cloudsketch/internal/providers"
	"cloudsketch/internal/providers/azure/containers/management_group"
	"cloudsketch/internal/providers/azure/containers/subscription"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"fmt"
	"log"
	"strings"

	domainTypes "cloudsketch/internal/frontends/types"
	azContext "cloudsketch/internal/providers/azure/context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type azureProvider struct{}

func NewProvider() *azureProvider {
	return &azureProvider{}
}

func (h *azureProvider) FetchResources(input string) ([]*providers.Resource, string, error) {
	filenameWithSuffix := fmt.Sprintf("%s.json", input)

	cachedResources, ok := marshall.UnmarshalIfExists[[]*models.Resource](filenameWithSuffix)

	if ok {
		log.Printf("using existing file %s\n", filenameWithSuffix)

		return mapToProviderModel(*cachedResources), input, nil
	}

	credentials, err := azidentity.NewDefaultAzureCredential(nil)

	if err != nil {
		return nil, "", fmt.Errorf("authentication failure: %+v", err)
	}

	resources, tenantId, err := getResources(credentials, input)

	if err != nil {
		return nil, "", err
	}

	resources = normalize(resources, tenantId)

	// input resources can contain references to resources that do not exist (in other subscriptions for example). These need to be removed
	resources = filterUnknownDependencies(resources)

	return mapToProviderModel(resources), input, nil
}

func getResources(credentials *azidentity.DefaultAzureCredential, input string) ([]*models.Resource, string, error) {
	ctx := &azContext.Context{
		Resource: &azContext.ResourceIdentifier{
			Id: input,
		},
		Credentials: credentials,
	}

	if guid.IsGuid(input) {
		// subscription id
		return subscription.New().GetResource(ctx)
	}

	// management group
	return management_group.New().GetResource(ctx)
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

func normalize(resources []*models.Resource, tenantId string) []*models.Resource {
	unhandledTypes := set.New[string]()

	return list.Map(resources, func(resource *models.Resource) *models.Resource {
		return &models.Resource{
			Id:         strings.ToLower(resource.Id), // Azure is not consistent regarding casing. Ensure all id's are lowercase
			Type:       mapTypeToDomainType(resource.Type, unhandledTypes),
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

func generateAzurePortalLink(resource *models.Resource, tenant string) string {
	// https://portal.azure.com/#@<tenant>/resource/<resource id>
	return fmt.Sprintf("https://portal.azure.com/#@%s/resource%s", tenant, resource.Id)
}

func mapTypeToDomainType(azType string, unhandledTypes *set.Set[string]) string {
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
		types.MANAGED_CLUSTER:                       domainTypes.MANAGED_CLUSTER,
		types.MANAGEMENT_GROUP:                      domainTypes.MANAGEMENT_GROUP,
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
		seenResourceType := unhandledTypes.Contains(azType)

		// mechanism to prevent spamming the output with the same type
		if !seenResourceType {
			log.Printf("undefined mapping from Azure types %s to domain type", azType)
			unhandledTypes.Add(azType)
		}

		return azType
	}

	return domainType
}

func filterUnknownDependencies(resources []*models.Resource) []*models.Resource {
	for _, resource := range resources {
		dependenciesWithUnknownResourcesRemoved := list.Filter(resource.DependsOn, func(resourceDependencyId string) bool {
			targetDependency := list.FirstOrDefault(resources, nil, func(target *models.Resource) bool {
				return target.Id == resourceDependencyId
			})

			// a resource has a depedency to another resource which is not known
			if targetDependency == nil {
				log.Printf("removed unknown resource %s\n", resourceDependencyId)
				return false
			}

			return true
		})

		resource.DependsOn = dependenciesWithUnknownResourcesRemoved
	}

	return resources
}
