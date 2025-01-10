package main

import (
	"azsample/internal/az"
	"azsample/internal/drawio"
	"azsample/internal/handlers/application_gateway"
	"azsample/internal/handlers/data_factory"
	"azsample/internal/handlers/function_app"
	"azsample/internal/handlers/load_balancer"
	"azsample/internal/handlers/network_interface"
	"azsample/internal/handlers/private_dns_zone"
	"azsample/internal/handlers/private_endpoint"
	"azsample/internal/handlers/private_link_service"
	"azsample/internal/handlers/public_ip_address"
	"azsample/internal/handlers/resource_group"
	"azsample/internal/handlers/subscription"
	"azsample/internal/handlers/virtual_machine_scale_set"
	"azsample/internal/handlers/virtual_network"
	"azsample/internal/list"
	"encoding/json"
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
		az.FUNCTION_APP:              function_app.New().Handle,
		az.LOAD_BALANCER:             load_balancer.New().Handle,
		az.NETWORK_INTERFACE:         network_interface.New().Handle,
		az.PRIVATE_DNS_ZONE:          private_dns_zone.New().Handle,
		az.PRIVATE_ENDPOINT:          private_endpoint.New().Handle,
		az.PRIVATE_LINK_SERVICE:      private_link_service.New().Handle,
		az.PUBLIC_IP_ADDRESS:         public_ip_address.New().Handle,
		az.VIRTUAL_MACHINE_SCALE_SET: virtual_machine_scale_set.New().Handle,
		az.VIRTUAL_NETWORK:           virtual_network.New().Handle,
	}
	useFile = false
)

func main() {
	if useFile {
		log.Println("using existing file")

		resources := unmarshallResources()

		drawio.New().WriteDiagram("./out.drawio", resources)

		return
	}

	args := os.Args

	if len(args) < 2 {
		log.Fatalf("missing subscriptionId")
		return
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("authentication failure: %+v", err)
	}

	appContext = &az.Context{
		SubscriptionId: args[1],
		Credentials:    cred,
	}

	resourceGroups, err := subscription.New().Handle(appContext.SubscriptionId, appContext.Credentials)

	if err != nil {
		log.Fatalf("listing of resource groups failed: %+v", err)
	}

	resources := list.FlatMap(resourceGroups, func(resourceGroup *armresources.ResourceGroup) []*az.Resource {
		resource, err := resource_group.New().Handle(&az.Context{
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

	resources = list.FlatMap(resources, func(resource *az.Resource) []*az.Resource {
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

	for _, r := range resources {
		r.Id = strings.ToLower(r.Id)
		r.DependsOn = list.Map(r.DependsOn, strings.ToLower)
	}

	marshallResources(resources)

	drawio.New().WriteDiagram("./out.drawio", resources)
}

func marshallResources(resources []*az.Resource) {
	bytes, err := json.Marshal(resources)

	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("./example.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	file.Write(bytes)
}

func unmarshallResources() []*az.Resource {
	bytes, err := os.ReadFile("./example.txt")

	if err != nil {
		log.Fatal(err)
	}

	var resources []*az.Resource

	json.Unmarshal(bytes, &resources)

	return resources
}
