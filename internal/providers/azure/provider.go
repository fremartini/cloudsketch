package azure

import (
	"cloudsketch/internal/list"
	"cloudsketch/internal/marshall"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/handlers/application_gateway"
	"cloudsketch/internal/providers/azure/handlers/data_factory"
	"cloudsketch/internal/providers/azure/handlers/load_balancer"
	"cloudsketch/internal/providers/azure/handlers/nat_gateway"
	"cloudsketch/internal/providers/azure/handlers/network_interface"
	"cloudsketch/internal/providers/azure/handlers/private_dns_zone"
	"cloudsketch/internal/providers/azure/handlers/private_endpoint"
	"cloudsketch/internal/providers/azure/handlers/private_link_service"
	"cloudsketch/internal/providers/azure/handlers/public_ip_address"
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

type handleFunc = func(*azContext.Context) ([]*models.Resource, error)

var (
	handlers map[string]handleFunc = map[string]handleFunc{
		types.APPLICATION_GATEWAY:       application_gateway.New().Handle,
		types.DATA_FACTORY:              data_factory.New().Handle,
		types.LOAD_BALANCER:             load_balancer.New().Handle,
		types.NAT_GATEWAY:               nat_gateway.New().Handle,
		types.NETWORK_INTERFACE:         network_interface.New().Handle,
		types.PRIVATE_DNS_ZONE:          private_dns_zone.New().Handle,
		types.PRIVATE_ENDPOINT:          private_endpoint.New().Handle,
		types.PRIVATE_LINK_SERVICE:      private_link_service.New().Handle,
		types.PUBLIC_IP_ADDRESS:         public_ip_address.New().Handle,
		types.VIRTUAL_MACHINE:           virtual_machine.New().Handle,
		types.VIRTUAL_MACHINE_SCALE_SET: virtual_machine_scale_set.New().Handle,
		types.VIRTUAL_NETWORK:           virtual_network.New().Handle,
		types.WEB_SITES:                 web_sites.New().Handle,
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

	cachedResources, ok := marshall.UnmarshalIfExists(filenameWithSuffix)

	if ok {
		log.Printf("using existing file %s\n", filenameWithSuffix)

		return cachedResources, filename, nil
	}

	allResources, err := resource_group.New().Handle(ctx)

	if err != nil {
		return nil, "", err
	}

	// add the subscription entry
	allResources = append(allResources, &models.Resource{
		Id:   subscription.Id,
		Name: subscription.Name,
		Type: types.SUBSCRIPTION,
	})

	allResources = list.FlatMap(allResources, func(resource *models.Resource) []*models.Resource {
		log.Print(resource.Name)

		f, ok := handlers[resource.Type]

		// no handler is registered. Add the resource as-is
		if !ok {
			return []*models.Resource{resource}
		}

		// handler is registered. Add whatever it returns
		resourcesToAdd, err := f(&azContext.Context{
			SubscriptionId:    ctx.SubscriptionId,
			TenantId:          ctx.TenantId,
			Credentials:       ctx.Credentials,
			ResourceGroupName: resource.ResourceGroup,
			ResourceName:      resource.Name,
			ResourceId:        resource.Id,
		})

		if err != nil {
			log.Fatal(err)
		}

		return resourcesToAdd
	})

	// ensure all id's are lowercase and map Azure types to domain types
	for _, r := range allResources {
		r.Id = strings.ToLower(r.Id)
		r.DependsOn = list.Map(r.DependsOn, strings.ToLower)
	}

	domainModels := list.Map(allResources, func(r *models.Resource) *domainModels.Resource {
		return mapToDomainResource(r, ctx.TenantId)
	})

	// cache resources for next run
	err = marshall.MarshallResources(filenameWithSuffix, domainModels)

	if err != nil {
		return nil, "", err
	}

	return domainModels, filename, nil
}

func mapToDomainResource(resource *models.Resource, tenantId string) *domainModels.Resource {
	properties := resource.Properties

	if properties == nil {
		properties = map[string]string{}
	}

	properties["link"] = generateAzurePortalLink(resource, tenantId)

	return &domainModels.Resource{
		Id:         resource.Id,
		Type:       mapTypeToDomainType(resource.Type),
		Name:       resource.Name,
		DependsOn:  resource.DependsOn,
		Properties: properties,
	}
}

func generateAzurePortalLink(resource *models.Resource, tenant string) string {
	// https://portal.azure.com/#@<tenant>/resource/<resource id>
	return fmt.Sprintf("https://portal.azure.com/#@%s/resource%s", tenant, resource.Id)
}

func mapTypeToDomainType(azType string) string {
	domainTypes := map[string]string{
		types.APP_SERVICE:                           domainTypes.APP_SERVICE,
		types.APP_SERVICE_PLAN:                      domainTypes.APP_SERVICE_PLAN,
		types.APPLICATION_GATEWAY:                   domainTypes.APPLICATION_GATEWAY,
		types.APPLICATION_INSIGHTS:                  domainTypes.APPLICATION_INSIGHTS,
		types.APPLICATION_SECURITY_GROUP:            domainTypes.APPLICATION_SECURITY_GROUP,
		types.CONTAINER_REGISTRY:                    domainTypes.CONTAINER_REGISTRY,
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
		types.NAT_GATEWAY:                           domainTypes.NAT_GATEWAY,
		types.NETWORK_INTERFACE:                     domainTypes.NETWORK_INTERFACE,
		types.NETWORK_SECURITY_GROUP:                domainTypes.NETWORK_SECURITY_GROUP,
		types.POSTGRES_SQL_SERVER:                   domainTypes.POSTGRES_SQL_SERVER,
		types.PRIVATE_DNS_ZONE:                      domainTypes.PRIVATE_DNS_ZONE,
		types.PRIVATE_ENDPOINT:                      domainTypes.PRIVATE_ENDPOINT,
		types.PRIVATE_LINK_SERVICE:                  domainTypes.PRIVATE_LINK_SERVICE,
		types.PUBLIC_IP_ADDRESS:                     domainTypes.PUBLIC_IP_ADDRESS,
		types.ROUTE_TABLE:                           domainTypes.ROUTE_TABLE,
		types.SQL_DATABASE:                          domainTypes.SQL_DATABASE,
		types.SQL_SERVER:                            domainTypes.SQL_SERVER,
		types.STORAGE_ACCOUNT:                       domainTypes.STORAGE_ACCOUNT,
		types.SUBNET:                                domainTypes.SUBNET,
		types.SUBSCRIPTION:                          domainTypes.SUBSCRIPTION,
		types.VIRTUAL_MACHINE:                       domainTypes.VIRTUAL_MACHINE,
		types.VIRTUAL_MACHINE_SCALE_SET:             domainTypes.VIRTUAL_MACHINE_SCALE_SET,
		types.VIRTUAL_NETWORK:                       domainTypes.VIRTUAL_NETWORK,
	}

	domainType, ok := domainTypes[azType]

	if !ok {
		log.Printf("undefined mapping from Azure types %s to domain type", azType)
		return azType
	}

	return domainType
}
