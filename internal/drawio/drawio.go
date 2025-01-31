package drawio

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/app_service_plan"
	"azsample/internal/drawio/handlers/application_gateway"
	"azsample/internal/drawio/handlers/container_registry"
	"azsample/internal/drawio/handlers/data_factory"
	"azsample/internal/drawio/handlers/data_factory_integration_runtime"
	"azsample/internal/drawio/handlers/data_factory_managed_private_endpoint"
	"azsample/internal/drawio/handlers/databricks_workspace"
	"azsample/internal/drawio/handlers/diagram"
	"azsample/internal/drawio/handlers/dns_record"
	"azsample/internal/drawio/handlers/key_vault"
	"azsample/internal/drawio/handlers/load_balancer"
	"azsample/internal/drawio/handlers/load_balancer_frontend"
	"azsample/internal/drawio/handlers/nat_gateway"
	"azsample/internal/drawio/handlers/network_interface"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/handlers/postgres_sql_server"
	"azsample/internal/drawio/handlers/private_dns_zone"
	"azsample/internal/drawio/handlers/private_endpoint"
	"azsample/internal/drawio/handlers/private_link_service"
	"azsample/internal/drawio/handlers/public_ip_address"
	"azsample/internal/drawio/handlers/sql_database"
	"azsample/internal/drawio/handlers/sql_server"
	"azsample/internal/drawio/handlers/storage_account"
	"azsample/internal/drawio/handlers/subnet"
	"azsample/internal/drawio/handlers/virtual_machine"
	"azsample/internal/drawio/handlers/virtual_machine_scale_set"
	"azsample/internal/drawio/handlers/virtual_network"
	"azsample/internal/drawio/handlers/web_sites"
	"azsample/internal/list"
	"fmt"
	"log"
)

type handleFuncMap = map[string]handler

type handler interface {
	DrawIcon(*az.Resource, *map[string]*node.ResourceAndNode) []*node.Node
	DrawDependency(*az.Resource, *az.Resource, *map[string]*node.ResourceAndNode) *node.Arrow
	DrawBox(*az.Resource, []*az.Resource, *map[string]*node.ResourceAndNode) []*node.Node
}

var (
	commands handleFuncMap = handleFuncMap{
		app_service_plan.TYPE:    app_service_plan.New(),
		application_gateway.TYPE: application_gateway.New(),
		//application_insights.TYPE:                  application_insights.New(),
		//application_security_group.TYPE:            application_security_group.New(),
		container_registry.TYPE:                    container_registry.New(),
		data_factory.TYPE:                          data_factory.New(),
		data_factory_integration_runtime.TYPE:      data_factory_integration_runtime.New(),
		data_factory_managed_private_endpoint.TYPE: data_factory_managed_private_endpoint.New(),
		databricks_workspace.TYPE:                  databricks_workspace.New(),
		dns_record.TYPE:                            dns_record.New(),
		key_vault.TYPE:                             key_vault.New(),
		load_balancer.TYPE:                         load_balancer.New(),
		load_balancer_frontend.TYPE:                load_balancer_frontend.New(),
		//log_analytics.TYPE:                         log_analytics.New(),
		nat_gateway.TYPE:       nat_gateway.New(),
		network_interface.TYPE: network_interface.New(),
		//network_security_group.TYPE:                network_security_group.New(),
		postgres_sql_server.TYPE:       postgres_sql_server.New(),
		private_dns_zone.TYPE:          private_dns_zone.New(),
		private_endpoint.TYPE:          private_endpoint.New(),
		private_link_service.TYPE:      private_link_service.New(),
		public_ip_address.TYPE:         public_ip_address.New(),
		sql_database.TYPE:              sql_database.New(),
		sql_server.TYPE:                sql_server.New(),
		storage_account.TYPE:           storage_account.New(),
		subnet.TYPE:                    subnet.New(),
		virtual_machine.TYPE:           virtual_machine.New(),
		virtual_machine_scale_set.TYPE: virtual_machine_scale_set.New(),
		virtual_network.TYPE:           virtual_network.New(),
		web_sites.TYPE:                 web_sites.New(),
	}
	resource_map             = map[string]*node.ResourceAndNode{}
	seen_unhandled_resources = map[string]bool{}
)

type drawio struct {
}

func New() *drawio {
	return &drawio{}
}

func (d *drawio) WriteDiagram(filename string, resources []*az.Resource) {
	// at this point only the Azure resources are known - this function adds the corresponding DrawIO icons
	cells := addDrawIOCells(resources)

	// some resources like vnets and subnets needs boxes draw around them, and their resources moved into them
	boxes := addBoxes()

	// with every DrawIO icon present, add the dependency arrows
	dependencyArrows := addDependencyArrows()

	// combine everything and render them in the final diagram
	// items appended first are rendered first (in the background)
	cellsToRender := []string{}
	cellsToRender = append(cellsToRender, boxes...)
	cellsToRender = append(cellsToRender, dependencyArrows...)
	cellsToRender = append(cellsToRender, list.Map(cells, node.ToMXCell)...)

	dgrm := diagram.New(cellsToRender)

	if err := dgrm.Write(filename); err != nil {
		log.Fatal(err)
	}
}

func addDrawIOCells(resources []*az.Resource) []*node.Node {
	var drawables []*node.Node
	for _, resource := range resources {
		// draw dependencies
		deps := drawDependenciesRecursively(resource, resources)

		drawables = append(drawables, deps...)

		if resource_map[resource.Id] != nil {
			// resource already drawn
			continue
		}

		// draw this resource
		nodes := drawResource(resource)

		drawables = append(drawables, nodes...)
	}

	return drawables
}

func drawDependenciesRecursively(resource *az.Resource, resources []*az.Resource) []*node.Node {
	var drawables []*node.Node
	for _, dependencyId := range resource.DependsOn {
		if resource_map[dependencyId] != nil {
			// dependency already drawn
			continue
		}

		dependency := list.First(resources, func(r *az.Resource) bool {
			return r.Id == dependencyId
		})

		dependencyNodes := drawDependenciesRecursively(dependency, resources)
		nodes := drawResource(dependency)

		drawables = append(drawables, dependencyNodes...)
		drawables = append(drawables, nodes...)
	}
	return drawables
}

func drawResource(resource *az.Resource) []*node.Node {
	f, ok := commands[resource.Type]

	if !ok {
		_, seenResource := seen_unhandled_resources[resource.Type]

		// mechanism to prevent spamming the output with the same type
		if !seenResource {
			log.Printf("unhandled type %s", resource.Type)
			seen_unhandled_resources[resource.Type] = true
		}

		return []*node.Node{}
	}

	nodes := f.DrawIcon(resource, &resource_map)

	resource_map[resource.Id] = &node.ResourceAndNode{
		Resource: resource,
	}

	// some icons should not be rendered (duplicate private endpoints on Storage Accounts)
	if len(nodes) == 0 {
		delete(resource_map, resource.Id)
	} else {
		// prioritize the last element
		resource_map[resource.Id].Node = nodes[len(nodes)-1]
	}

	return nodes
}

func addDependencyArrows() []string {
	var cells []string

	for _, resourceAndNode := range resource_map {
		resource := resourceAndNode.Resource

		for _, dependency := range resource.DependsOn {
			f, ok := commands[resource.Type]

			if !ok {
				log.Fatalf("type %s has not been registered for rendering", resource.Type)
			}

			ok = resource_map[resource.Id].Node != nil

			if !ok {
				log.Printf("source %s was not drawn, skipping ...", resource.Id)
				continue
			}

			if _, ok := resource_map[dependency]; !ok {
				panic(fmt.Sprintf("%s not seen", dependency))
			}

			ok = resource_map[dependency].Node != nil

			if !ok {
				log.Printf("target %s was not drawn, skipping ...", dependency)
				continue
			}

			target := resource_map[dependency].Resource

			dependency := f.DrawDependency(resource, target, &resource_map)

			// dependency arrow may be omitted
			if dependency == nil {
				continue
			}

			cells = append(cells, dependency.ToMXCell())
		}
	}

	return cells
}

func addBoxes() []string {
	resources := []*az.Resource{}

	for _, resourceAndNode := range resource_map {
		resources = append(resources, resourceAndNode.Resource)
	}

	// TODO: implement a cleaner solution
	aspNodes := drawBoxForResourceType(resources, az.APP_SERVICE_PLAN)
	adfNodes := drawBoxForResourceType(resources, az.DATA_FACTORY)
	privateDNSZoneNodes := drawBoxForResourceType(resources, az.PRIVATE_DNS_ZONE)
	subnetNodes := drawBoxForResourceType(resources, az.SUBNET)
	vnetNodes := drawBoxForResourceType(resources, az.VIRTUAL_NETWORK)

	subnetCells := list.Map(subnetNodes, node.ToMXCell)
	vnetCells := list.Map(vnetNodes, node.ToMXCell)
	aspCells := list.Map(aspNodes, node.ToMXCell)
	adfCells := list.Map(adfNodes, node.ToMXCell)
	privateDNSZoneCells := list.Map(privateDNSZoneNodes, node.ToMXCell)

	// return vnets first so they are rendered in the background
	nodes := append(vnetCells, subnetCells...)
	nodes = append(nodes, aspCells...)
	nodes = append(nodes, adfCells...)
	nodes = append(nodes, privateDNSZoneCells...)

	return nodes
}

func drawBoxForResourceType(resources []*az.Resource, typ string) []*node.Node {
	nodes := []*node.Node{}

	resourcesWithType := list.Filter(resources, func(r *az.Resource) bool {
		return r.Type == typ
	})

	for _, resource := range resourcesWithType {
		nodes = append(nodes, commands[typ].DrawBox(resource, resources, &resource_map)...)
	}

	return nodes
}
