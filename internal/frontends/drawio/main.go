package drawio

import (
	"cloudsketch/internal/config"
	"cloudsketch/internal/datastructures/build_graph"
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/frontends/drawio/handlers/ai_services"
	"cloudsketch/internal/frontends/drawio/handlers/app_service"
	"cloudsketch/internal/frontends/drawio/handlers/app_service_plan"
	"cloudsketch/internal/frontends/drawio/handlers/application_gateway"
	"cloudsketch/internal/frontends/drawio/handlers/application_group"
	"cloudsketch/internal/frontends/drawio/handlers/application_insights"
	"cloudsketch/internal/frontends/drawio/handlers/application_security_group"
	"cloudsketch/internal/frontends/drawio/handlers/bastion"
	"cloudsketch/internal/frontends/drawio/handlers/connection"
	"cloudsketch/internal/frontends/drawio/handlers/container_registry"
	"cloudsketch/internal/frontends/drawio/handlers/cosmos"
	"cloudsketch/internal/frontends/drawio/handlers/data_factory"
	"cloudsketch/internal/frontends/drawio/handlers/data_factory_integration_runtime"
	"cloudsketch/internal/frontends/drawio/handlers/data_factory_managed_private_endpoint"
	"cloudsketch/internal/frontends/drawio/handlers/databricks_workspace"
	"cloudsketch/internal/frontends/drawio/handlers/diagram"
	"cloudsketch/internal/frontends/drawio/handlers/dns_record"
	"cloudsketch/internal/frontends/drawio/handlers/express_route_circuit"
	"cloudsketch/internal/frontends/drawio/handlers/express_route_gateway"
	"cloudsketch/internal/frontends/drawio/handlers/function_app"
	"cloudsketch/internal/frontends/drawio/handlers/host_pool"
	"cloudsketch/internal/frontends/drawio/handlers/key_vault"
	"cloudsketch/internal/frontends/drawio/handlers/load_balancer"
	"cloudsketch/internal/frontends/drawio/handlers/load_balancer_frontend"
	"cloudsketch/internal/frontends/drawio/handlers/log_analytics"
	"cloudsketch/internal/frontends/drawio/handlers/logic_app"
	"cloudsketch/internal/frontends/drawio/handlers/machine_learning_workspace"
	"cloudsketch/internal/frontends/drawio/handlers/nat_gateway"
	"cloudsketch/internal/frontends/drawio/handlers/network_interface"
	"cloudsketch/internal/frontends/drawio/handlers/network_security_group"
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/handlers/postgres_sql_server"
	"cloudsketch/internal/frontends/drawio/handlers/private_dns_resolver"
	"cloudsketch/internal/frontends/drawio/handlers/private_dns_zone"
	"cloudsketch/internal/frontends/drawio/handlers/private_endpoint"
	"cloudsketch/internal/frontends/drawio/handlers/private_link_service"
	"cloudsketch/internal/frontends/drawio/handlers/public_ip_address"
	"cloudsketch/internal/frontends/drawio/handlers/recovery_service_vault"
	"cloudsketch/internal/frontends/drawio/handlers/redis"
	"cloudsketch/internal/frontends/drawio/handlers/route_table"
	"cloudsketch/internal/frontends/drawio/handlers/search_service"
	"cloudsketch/internal/frontends/drawio/handlers/signalr"
	"cloudsketch/internal/frontends/drawio/handlers/sql_database"
	"cloudsketch/internal/frontends/drawio/handlers/sql_server"
	"cloudsketch/internal/frontends/drawio/handlers/static_web_app"
	"cloudsketch/internal/frontends/drawio/handlers/storage_account"
	"cloudsketch/internal/frontends/drawio/handlers/subnet"
	"cloudsketch/internal/frontends/drawio/handlers/subscription"
	"cloudsketch/internal/frontends/drawio/handlers/user_assigned_identity"
	"cloudsketch/internal/frontends/drawio/handlers/virtual_hub"
	"cloudsketch/internal/frontends/drawio/handlers/virtual_machine"
	"cloudsketch/internal/frontends/drawio/handlers/virtual_machine_scale_set"
	"cloudsketch/internal/frontends/drawio/handlers/virtual_network"
	"cloudsketch/internal/frontends/drawio/handlers/virtual_network_gateway"
	"cloudsketch/internal/frontends/drawio/handlers/virtual_wan"
	"cloudsketch/internal/frontends/drawio/handlers/workspace"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"

	"cloudsketch/internal/list"
	"fmt"
	"log"
)

type handler interface {
	MapResource(*models.Resource) *node.Node
	PostProcessIcon(*node.ResourceAndNode, *map[string]*node.ResourceAndNode) *node.Node
	DrawDependencies(*models.Resource, []*models.Resource, *map[string]*node.ResourceAndNode) []*node.Arrow
	GroupResources(*models.Resource, []*models.Resource, *map[string]*node.ResourceAndNode) []*node.Node
}

var (
	commands map[string]handler = map[string]handler{
		ai_services.TYPE:                           ai_services.New(),
		app_service.TYPE:                           app_service.New(),
		app_service_plan.TYPE:                      app_service_plan.New(),
		application_gateway.TYPE:                   application_gateway.New(),
		application_group.TYPE:                     application_group.New(),
		application_insights.TYPE:                  application_insights.New(),
		application_security_group.TYPE:            application_security_group.New(),
		bastion.TYPE:                               bastion.New(),
		connection.TYPE:                            connection.New(),
		container_registry.TYPE:                    container_registry.New(),
		cosmos.TYPE:                                cosmos.New(),
		data_factory.TYPE:                          data_factory.New(),
		data_factory_integration_runtime.TYPE:      data_factory_integration_runtime.New(),
		data_factory_managed_private_endpoint.TYPE: data_factory_managed_private_endpoint.New(),
		databricks_workspace.TYPE:                  databricks_workspace.New(),
		dns_record.TYPE:                            dns_record.New(),
		express_route_circuit.TYPE:                 express_route_circuit.New(),
		express_route_gateway.TYPE:                 express_route_gateway.New(),
		function_app.TYPE:                          function_app.New(),
		host_pool.TYPE:                             host_pool.New(),
		key_vault.TYPE:                             key_vault.New(),
		load_balancer.TYPE:                         load_balancer.New(),
		load_balancer_frontend.TYPE:                load_balancer_frontend.New(),
		log_analytics.TYPE:                         log_analytics.New(),
		logic_app.TYPE:                             logic_app.New(),
		machine_learning_workspace.TYPE:            machine_learning_workspace.New(),
		nat_gateway.TYPE:                           nat_gateway.New(),
		network_interface.TYPE:                     network_interface.New(),
		network_security_group.TYPE:                network_security_group.New(),
		postgres_sql_server.TYPE:                   postgres_sql_server.New(),
		private_dns_resolver.TYPE:                  private_dns_resolver.New(),
		private_dns_zone.TYPE:                      private_dns_zone.New(),
		private_endpoint.TYPE:                      private_endpoint.New(),
		private_link_service.TYPE:                  private_link_service.New(),
		public_ip_address.TYPE:                     public_ip_address.New(),
		recovery_service_vault.TYPE:                recovery_service_vault.New(),
		redis.TYPE:                                 redis.New(),
		route_table.TYPE:                           route_table.New(),
		search_service.TYPE:                        search_service.New(),
		signalr.TYPE:                               signalr.New(),
		sql_database.TYPE:                          sql_database.New(),
		sql_server.TYPE:                            sql_server.New(),
		static_web_app.TYPE:                        static_web_app.New(),
		storage_account.TYPE:                       storage_account.New(),
		subnet.TYPE:                                subnet.New(),
		subscription.TYPE:                          subscription.New(),
		user_assigned_identity.TYPE:                user_assigned_identity.New(),
		virtual_hub.TYPE:                           virtual_hub.New(),
		virtual_machine.TYPE:                       virtual_machine.New(),
		virtual_machine_scale_set.TYPE:             virtual_machine_scale_set.New(),
		virtual_network.TYPE:                       virtual_network.New(),
		virtual_network_gateway.TYPE:               virtual_network_gateway.New(),
		virtual_wan.TYPE:                           virtual_wan.New(),
		workspace.TYPE:                             workspace.New(),
	}
)

type drawio struct {
}

func New() *drawio {
	return &drawio{}
}

func removeBlacklistedHandlers() {
	config, ok := config.Read()

	if !ok {
		return
	}

	// remove entries on the blacklist
	for _, blacklistedItem := range config.Blacklist {
		delete(commands, blacklistedItem)
	}
}

func (d *drawio) WriteDiagram(resources []*models.Resource, filename string) error {
	removeBlacklistedHandlers()

	// at this point only the Azure resources are known - this function adds the corresponding DrawIO icons
	resource_map, err := populateResourceMap(resources)

	if err != nil {
		return err
	}

	// some resources group other resources
	groups := postProcessIcons(resource_map)

	// some resources like vnets and subnets needs boxes draw around them, and their resources moved into them
	boxes := groupResources(resource_map)

	// with every DrawIO icon present, add the dependency arrows
	dependencyArrows := addDependencies(resource_map)

	allResources := []*node.ResourceAndNode{}
	for _, resource := range *resource_map {
		allResources = append(allResources, resource)
	}

	// private endpoints, NICs, PIPs and NSGs are typically used as icons attached to other icons and should therefore be rendered in front of them
	overlayResources := []string{types.PRIVATE_ENDPOINT, types.NETWORK_INTERFACE, types.PUBLIC_IP_ADDRESS, types.NETWORK_SECURITY_GROUP, types.ROUTE_TABLE}
	allResourcesThatShouldGoInFront, allResourcesThatShouldGoInBack := list.Split(allResources, func(n *node.ResourceAndNode) bool {
		return list.Contains(overlayResources, func(typ string) bool {
			return n.Resource.Type == typ
		})
	})

	allResources = append(allResourcesThatShouldGoInBack, allResourcesThatShouldGoInFront...)
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

func populateResourceMap(resources []*models.Resource) (*map[string]*node.ResourceAndNode, error) {
	resource_map := &map[string]*node.ResourceAndNode{}
	unhandled_resources := set.New[string]()

	tasks := list.Map(resources, func(r *models.Resource) *build_graph.Task {
		return build_graph.NewTask(r.Id, list.Map(r.DependsOn, func(m *models.Resource) string { return m.Id }), []string{}, []string{}, func() { drawResource(r, unhandled_resources, resource_map) })
	})

	bg, err := build_graph.NewGraph(tasks)

	if err != nil {
		return nil, fmt.Errorf("error during construction of dependency graph: %+v", err)
	}

	// ensure all resources that depend on this have been draw
	for _, task := range tasks {
		bg.ResolveInverse(task)
	}

	return resource_map, nil
}

func drawResource(resource *models.Resource, unhandled_resources *set.Set[string], resource_map *map[string]*node.ResourceAndNode) {
	if (*resource_map)[resource.Id] != nil {
		// resource already drawn
		return
	}

	f, ok := commands[resource.Type]

	if !ok {
		seenResourceType := unhandled_resources.Contains(resource.Type)

		// mechanism to prevent spamming the output with the same type
		if !seenResourceType {
			log.Printf("unhandled type %s", resource.Type)
			unhandled_resources.Add(resource.Type)
		}

		return
	}

	icon := f.MapResource(resource)

	(*resource_map)[resource.Id] = &node.ResourceAndNode{
		Resource: resource,
		Node:     icon,
	}
}

func postProcessIcons(resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	for _, resource := range *resource_map {
		nodeToAdd := commands[resource.Resource.Type].PostProcessIcon(resource, resource_map)

		if nodeToAdd == nil {
			continue
		}

		nodes = append(nodes, nodeToAdd)
	}

	return nodes
}

func addDependencies(resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	var arrows []*node.Arrow

	for _, resourceAndNode := range *resource_map {
		resource := resourceAndNode.Resource

		f, ok := commands[resource.Type]

		if !ok {
			log.Fatalf("type %s has not been registered for rendering", resource.Type)
		}

		dependencyIds := list.Filter(resource.DependsOn, func(dependency *models.Resource) bool {
			targetMissing := (*resource_map)[dependency.Id] == nil || (*resource_map)[dependency.Id].Node == nil
			if targetMissing {
				log.Printf("target %s was not drawn, skipping", dependency.Id)
				return false
			}

			return true
		})

		resources := list.Map(dependencyIds, func(dependency *models.Resource) *models.Resource {
			return (*resource_map)[dependency.Id].Resource
		})

		arrowsToAdd := f.DrawDependencies(resource, resources, resource_map)

		arrows = append(arrows, arrowsToAdd...)
	}

	return arrows
}

func groupResources(resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resources := []*models.Resource{}

	for _, resourceAndNode := range *resource_map {
		resources = append(resources, resourceAndNode.Resource)
	}

	resourcesWithoutVnetsAndSubnets := list.Filter(resources, func(resource *models.Resource) bool {
		return resource.Type != types.SUBNET && resource.Type != types.VIRTUAL_NETWORK && resource.Type != types.SUBSCRIPTION
	})

	boxes := list.FlatMap(resourcesWithoutVnetsAndSubnets, func(resource *models.Resource) []*node.Node {
		return commands[resource.Type].GroupResources(resource, resources, resource_map)
	})

	// virtual netwoks, subnets and subscription needs to be handled last since they "depend" on all other resources
	subnets := drawGroupForResourceType(resources, types.SUBNET, resource_map)
	vnets := drawGroupForResourceType(resources, types.VIRTUAL_NETWORK, resource_map)
	subscriptions := drawGroupForResourceType(resources, types.SUBSCRIPTION, resource_map)

	// return subscriptions first so they are rendered in the background
	nodes := append(subscriptions, append(vnets, append(subnets, boxes...)...)...)

	return nodes
}

func drawGroupForResourceType(resources []*models.Resource, typ string, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesWithType := list.Filter(resources, func(r *models.Resource) bool {
		return r.Type == typ
	})

	nodes := list.FlatMap(resourcesWithType, func(resource *models.Resource) []*node.Node {
		return commands[typ].GroupResources(resource, resources, resource_map)
	})

	return nodes
}
