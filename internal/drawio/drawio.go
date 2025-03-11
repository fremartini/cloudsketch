package drawio

import (
	"cloudsketch/internal/datastructures/build_graph"
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/drawio/handlers/app_service"
	"cloudsketch/internal/drawio/handlers/app_service_plan"
	"cloudsketch/internal/drawio/handlers/application_gateway"
	"cloudsketch/internal/drawio/handlers/application_insights"
	"cloudsketch/internal/drawio/handlers/application_security_group"
	"cloudsketch/internal/drawio/handlers/container_registry"
	"cloudsketch/internal/drawio/handlers/data_factory"
	"cloudsketch/internal/drawio/handlers/data_factory_integration_runtime"
	"cloudsketch/internal/drawio/handlers/data_factory_managed_private_endpoint"
	"cloudsketch/internal/drawio/handlers/databricks_workspace"
	"cloudsketch/internal/drawio/handlers/diagram"
	"cloudsketch/internal/drawio/handlers/dns_record"
	"cloudsketch/internal/drawio/handlers/function_app"
	"cloudsketch/internal/drawio/handlers/key_vault"
	"cloudsketch/internal/drawio/handlers/load_balancer"
	"cloudsketch/internal/drawio/handlers/load_balancer_frontend"
	"cloudsketch/internal/drawio/handlers/log_analytics"
	"cloudsketch/internal/drawio/handlers/logic_app"
	"cloudsketch/internal/drawio/handlers/nat_gateway"
	"cloudsketch/internal/drawio/handlers/network_interface"
	"cloudsketch/internal/drawio/handlers/network_security_group"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/handlers/postgres_sql_server"
	"cloudsketch/internal/drawio/handlers/private_dns_zone"
	"cloudsketch/internal/drawio/handlers/private_endpoint"
	"cloudsketch/internal/drawio/handlers/private_link_service"
	"cloudsketch/internal/drawio/handlers/public_ip_address"
	"cloudsketch/internal/drawio/handlers/redis"
	"cloudsketch/internal/drawio/handlers/route_table"
	"cloudsketch/internal/drawio/handlers/sql_database"
	"cloudsketch/internal/drawio/handlers/sql_server"
	"cloudsketch/internal/drawio/handlers/storage_account"
	"cloudsketch/internal/drawio/handlers/subnet"
	"cloudsketch/internal/drawio/handlers/subscription"
	"cloudsketch/internal/drawio/handlers/virtual_machine"
	"cloudsketch/internal/drawio/handlers/virtual_machine_scale_set"
	"cloudsketch/internal/drawio/handlers/virtual_network"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
	"cloudsketch/internal/list"
	"fmt"
	"log"
)

type handleFuncMap = map[string]handler

type handler interface {
	MapResource(*models.Resource) *node.Node
	PostProcessIcon(*node.ResourceAndNode, *map[string]*node.ResourceAndNode) *node.Node
	DrawDependency(*models.Resource, []*models.Resource, *map[string]*node.ResourceAndNode) []*node.Arrow
	GroupResources(*models.Resource, []*models.Resource, *map[string]*node.ResourceAndNode) []*node.Node
}

var (
	commands handleFuncMap = handleFuncMap{
		app_service.TYPE:                           app_service.New(),
		app_service_plan.TYPE:                      app_service_plan.New(),
		application_gateway.TYPE:                   application_gateway.New(),
		application_insights.TYPE:                  application_insights.New(),
		application_security_group.TYPE:            application_security_group.New(),
		container_registry.TYPE:                    container_registry.New(),
		data_factory.TYPE:                          data_factory.New(),
		data_factory_integration_runtime.TYPE:      data_factory_integration_runtime.New(),
		data_factory_managed_private_endpoint.TYPE: data_factory_managed_private_endpoint.New(),
		databricks_workspace.TYPE:                  databricks_workspace.New(),
		dns_record.TYPE:                            dns_record.New(),
		function_app.TYPE:                          function_app.New(),
		key_vault.TYPE:                             key_vault.New(),
		load_balancer.TYPE:                         load_balancer.New(),
		load_balancer_frontend.TYPE:                load_balancer_frontend.New(),
		log_analytics.TYPE:                         log_analytics.New(),
		logic_app.TYPE:                             logic_app.New(),
		nat_gateway.TYPE:                           nat_gateway.New(),
		network_interface.TYPE:                     network_interface.New(),
		network_security_group.TYPE:                network_security_group.New(),
		postgres_sql_server.TYPE:                   postgres_sql_server.New(),
		private_dns_zone.TYPE:                      private_dns_zone.New(),
		private_endpoint.TYPE:                      private_endpoint.New(),
		private_link_service.TYPE:                  private_link_service.New(),
		public_ip_address.TYPE:                     public_ip_address.New(),
		redis.TYPE:                                 redis.New(),
		route_table.TYPE:                           route_table.New(),
		sql_database.TYPE:                          sql_database.New(),
		sql_server.TYPE:                            sql_server.New(),
		storage_account.TYPE:                       storage_account.New(),
		subnet.TYPE:                                subnet.New(),
		subscription.TYPE:                          subscription.New(),
		virtual_machine.TYPE:                       virtual_machine.New(),
		virtual_machine_scale_set.TYPE:             virtual_machine_scale_set.New(),
		virtual_network.TYPE:                       virtual_network.New(),
	}
)

type drawio struct {
}

func New() *drawio {
	return &drawio{}
}

func (d *drawio) WriteDiagram(filename string, resources []*models.Resource) error {
	// at this point only the Azure resources are known - this function adds the corresponding DrawIO icons
	resource_map, err := populateResourceMap(resources)

	if err != nil {
		return err
	}

	// some resources group other resources
	groups := processGroups(resource_map)

	// some resources like vnets and subnets needs boxes draw around them, and their resources moved into them
	boxes := addBoxes(resource_map)

	// with every DrawIO icon present, add the dependency arrows
	dependencyArrows := addDependencyArrows(resource_map)

	allResources := []*node.ResourceAndNode{}
	for _, resource := range *resource_map {
		allResources = append(allResources, resource)
	}

	// private endpoints and NICs are typically used as icons attached to other icons and should therefore be rendered in front
	// TODO: implement a better solution?
	allResourcesWithoutPEandNICs := list.Filter(allResources, func(n *node.ResourceAndNode) bool {
		return n.Resource.Type != types.PRIVATE_ENDPOINT && n.Resource.Type != types.NETWORK_INTERFACE
	})

	privateEndpointsAndNICS := list.Filter(allResources, func(n *node.ResourceAndNode) bool {
		return n.Resource.Type == types.PRIVATE_ENDPOINT || n.Resource.Type == types.NETWORK_INTERFACE
	})

	allResources = append(allResourcesWithoutPEandNICs, privateEndpointsAndNICS...)
	allResourcesNodes := list.Map(allResources, func(n *node.ResourceAndNode) *node.Node {
		return n.Node
	})

	// combine everything and render them in the final diagram
	// items appended first are rendered first (in the background)
	cellsToRender := []string{}
	cellsToRender = append(cellsToRender, list.Map(boxes, node.ToMXCell)...)
	cellsToRender = append(cellsToRender, list.Map(groups, node.ToMXCell)...)
	cellsToRender = append(cellsToRender, list.Map(dependencyArrows, func(a *node.Arrow) string {
		return a.ToMXCell()
	})...)
	cellsToRender = append(cellsToRender, list.Map(allResourcesNodes, node.ToMXCell)...)

	dgrm := diagram.New(cellsToRender)

	return dgrm.Write(filename)
}

func filterUnknownDependencies(resources []*models.Resource) []*models.Resource {
	for _, resource := range resources {
		resource.DependsOn = list.Filter(resource.DependsOn, func(d string) bool {
			dependency := list.FirstOrDefault(resources, nil, func(r *models.Resource) bool {
				return r.Id == d
			})

			return dependency != nil
		})
	}

	return resources
}

func populateResourceMap(resources []*models.Resource) (*map[string]*node.ResourceAndNode, error) {
	resource_map := &map[string]*node.ResourceAndNode{}
	seen_unhandled_resources := set.New[string]()

	// input resources can contain references to resources that do not exist (in other subscriptions for example). These need to be removed
	resources = filterUnknownDependencies(resources)

	tasks := list.Map(resources, func(r *models.Resource) *build_graph.Task {
		return build_graph.NewTask(r.Id, r.DependsOn, []string{}, []string{}, func() { drawResource(r, seen_unhandled_resources, resource_map) })
	})

	bg, err := build_graph.NewGraph(tasks)

	if err != nil {
		return nil, fmt.Errorf("error during construction of dependency graph: %+v", err)
	}

	for _, task := range tasks {
		bg.Resolve(task)
	}

	return resource_map, nil
}

func drawResource(resource *models.Resource, seen_unhandled_resources *set.Set[string], resource_map *map[string]*node.ResourceAndNode) {
	if (*resource_map)[resource.Id] != nil {
		// resource already drawn
		return
	}

	f, ok := commands[resource.Type]

	if !ok {
		seenResourceType := seen_unhandled_resources.Contains(resource.Type)

		// mechanism to prevent spamming the output with the same type
		if !seenResourceType {
			log.Printf("unhandled type %s", resource.Type)
			seen_unhandled_resources.Add(resource.Type)
		}

		return
	}

	icon := f.MapResource(resource)

	(*resource_map)[resource.Id] = &node.ResourceAndNode{
		Resource: resource,
		Node:     icon,
	}
}

func processGroups(resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	groups := []*node.Node{}

	for _, resource := range *resource_map {
		groupToAdd := commands[resource.Resource.Type].PostProcessIcon(resource, resource_map)

		if groupToAdd == nil {
			continue
		}

		groups = append(groups, groupToAdd)
	}
	return groups
}

func addDependencyArrows(resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	var nodes []*node.Arrow

	for _, resourceAndNode := range *resource_map {
		resource := resourceAndNode.Resource

		f, ok := commands[resource.Type]

		if !ok {
			log.Fatalf("type %s has not been registered for rendering", resource.Type)
		}

		dependencyIds := list.Filter(resource.DependsOn, func(dependency string) bool {
			targetMissing := (*resource_map)[dependency] == nil || (*resource_map)[dependency].Node == nil
			if targetMissing {
				log.Printf("target %s was not drawn, skipping ...", dependency)
				return false
			}

			return true
		})

		dependencyResources := list.Map(dependencyIds, func(dependency string) *models.Resource {
			return (*resource_map)[dependency].Resource
		})

		arrows := f.DrawDependency(resource, dependencyResources, resource_map)

		nodes = append(nodes, arrows...)
	}

	return nodes
}

func addBoxes(resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resources := []*models.Resource{}

	for _, resourceAndNode := range *resource_map {
		resources = append(resources, resourceAndNode.Resource)
	}

	resourcesWithoutVnetsAndSubnets := list.Filter(resources, func(resource *models.Resource) bool {
		return resource.Type != types.SUBNET && resource.Type != types.VIRTUAL_NETWORK
	})

	boxes := list.FlatMap(resourcesWithoutVnetsAndSubnets, func(resource *models.Resource) []*node.Node {
		return commands[resource.Type].GroupResources(resource, resources, resource_map)
	})

	// virtual netwoks and subnets needs to be handled last since they "depend" on all other resources
	subnets := DrawGroupForResourceType(resources, types.SUBNET, resource_map)
	vnets := DrawGroupForResourceType(resources, types.VIRTUAL_NETWORK, resource_map)

	// return vnets first so they are rendered in the background
	nodes := append(vnets, append(subnets, boxes...)...)

	return nodes
}

func DrawGroupForResourceType(resources []*models.Resource, typ string, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	resourcesWithType := list.Filter(resources, func(r *models.Resource) bool {
		return r.Type == typ
	})

	for _, resource := range resourcesWithType {
		nodes = append(nodes, commands[typ].GroupResources(resource, resources, resource_map)...)
	}

	return nodes
}
