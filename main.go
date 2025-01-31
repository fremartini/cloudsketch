package main

import (
	"azsample/internal/az"
	"azsample/internal/drawio"
	"azsample/internal/handlers/application_gateway"
	"azsample/internal/handlers/data_factory"
	"azsample/internal/handlers/load_balancer"
	"azsample/internal/handlers/nat_gateway"
	"azsample/internal/handlers/network_interface"
	"azsample/internal/handlers/private_dns_zone"
	"azsample/internal/handlers/private_endpoint"
	"azsample/internal/handlers/private_link_service"
	"azsample/internal/handlers/public_ip_address"
	"azsample/internal/handlers/resource_group"
	"azsample/internal/handlers/resources"
	"azsample/internal/handlers/subscription"
	"azsample/internal/handlers/virtual_machine_scale_set"
	"azsample/internal/handlers/virtual_network"
	"azsample/internal/handlers/web_sites"
	"azsample/internal/list"
	"azsample/marshall"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type handleFunc = func(*az.Context) ([]*az.Resource, error)

var (
	appContext *az.Context           = nil
	handlers   map[string]handleFunc = map[string]handleFunc{
		az.APPLICATION_GATEWAY:       application_gateway.New().Handle,
		az.DATA_FACTORY:              data_factory.New().Handle,
		az.LOAD_BALANCER:             load_balancer.New().Handle,
		az.NAT_GATEWAY:               nat_gateway.New().Handle,
		az.NETWORK_INTERFACE:         network_interface.New().Handle,
		az.PRIVATE_DNS_ZONE:          private_dns_zone.New().Handle,
		az.PRIVATE_ENDPOINT:          private_endpoint.New().Handle,
		az.PRIVATE_LINK_SERVICE:      private_link_service.New().Handle,
		az.PUBLIC_IP_ADDRESS:         public_ip_address.New().Handle,
		az.VIRTUAL_MACHINE_SCALE_SET: virtual_machine_scale_set.New().Handle,
		az.VIRTUAL_NETWORK:           virtual_network.New().Handle,
		az.WEB_SITES:                 web_sites.New().Handle,
	}
)

func main() {
	args := os.Args

	if len(args) < 2 {
		log.Fatalf("Missing Azure SubscriptionId")
		return
	}

	subscriptionId := args[1]

	// Fetch SubscriptionContext
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("authentication failure: %+v", err)
	}

	subscription := subscription.New().Handle(subscriptionId, cred)
	appContext = &az.Context{
		SubscriptionId: subscription.Id,
		Credentials:    cred,
	}

	filename := fmt.Sprintf("%s_%s.txt", subscription.Name, subscription.Id)

	canUseFile, allResources := marshall.UnmarshalIfExists(filename)

	if canUseFile {
		log.Println("using existing file")

		drawio.New().WriteDiagram(fmt.Sprintf("./%s.drawio", filename), allResources)

		return
	}

	resourceGroups, _ := resource_group.New().Handle(appContext.SubscriptionId, appContext.Credentials)
	if err != nil {
		log.Fatalf("listing of resource groups failed: %+v", err)
	}

	allResources = list.FlatMap(resourceGroups, func(resourceGroup *armresources.ResourceGroup) []*az.Resource {
		resource, err := resources.New().Handle(&az.Context{
			SubscriptionId: appContext.SubscriptionId,
			Credentials:    appContext.Credentials,
			ResourceName:   *resourceGroup.Name,
			ResourceGroup:  *resourceGroup.Name,
			ResourceId:     *resourceGroup.ID,
		})

		if err != nil {
			log.Fatal(err)
		}

		return resource
	})

	allResources = list.FlatMap(allResources, func(resource *az.Resource) []*az.Resource {
		log.Print(resource.Name)

		f, ok := handlers[resource.Type]

		if !ok {
			return []*az.Resource{{
				Id:            resource.Id,
				Name:          resource.Name,
				Type:          resource.Type,
				ResourceGroup: resource.ResourceGroup,
			}}
		}

		resources, err := f(&az.Context{
			SubscriptionId: appContext.SubscriptionId,
			Credentials:    appContext.Credentials,
			ResourceName:   resource.Name,
			ResourceGroup:  resource.ResourceGroup,
			ResourceId:     resource.Id,
		})

		if err != nil {
			log.Fatal(err)
		}

		return resources
	})

	for _, r := range allResources {
		r.Id = strings.ToLower(r.Id)
		r.DependsOn = list.Map(r.DependsOn, strings.ToLower)
	}

	marshall.MarshallResources(filename, allResources)

	drawio.New().WriteDiagram(fmt.Sprintf("./%s.drawio", filename), allResources)
}
