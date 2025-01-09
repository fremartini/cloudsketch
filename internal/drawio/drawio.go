package drawio

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/app_service_plan"
	"azsample/internal/drawio/handlers/data_factory"
	"azsample/internal/drawio/handlers/data_factory_integration_runtime"
	"azsample/internal/drawio/handlers/data_factory_managed_private_endpoint"
	"azsample/internal/drawio/handlers/dns_record"
	"azsample/internal/drawio/handlers/function_app"
	"azsample/internal/drawio/handlers/key_vault"
	"azsample/internal/drawio/handlers/load_balancer"
	"azsample/internal/drawio/handlers/load_balancer_frontend"
	"azsample/internal/drawio/handlers/nat_gateway"
	"azsample/internal/drawio/handlers/network_interface"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/handlers/private_dns_zone"
	"azsample/internal/drawio/handlers/private_endpoint"
	"azsample/internal/drawio/handlers/private_link_service"
	"azsample/internal/drawio/handlers/public_ip_address"
	"azsample/internal/drawio/handlers/sql_server"
	"azsample/internal/drawio/handlers/storage_account"
	"azsample/internal/drawio/handlers/subnet"
	"azsample/internal/drawio/handlers/virtual_machine"
	"azsample/internal/drawio/handlers/virtual_machine_scale_set"
	"azsample/internal/drawio/handlers/virtual_network"
	"azsample/internal/guid"
	"azsample/internal/list"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

type handleFuncMap = map[string]handler

type handler interface {
	DrawIcon(*az.Resource, *map[string]*node.ResourceAndNode) []*node.Node
	DrawDependency(*az.Resource, *az.Resource, *map[string]*node.Node) *node.Arrow
}

var (
	commands handleFuncMap = handleFuncMap{
		app_service_plan.TYPE: app_service_plan.New(),
		//application_insights.TYPE: application_insights.New(),
		//application_security_group.TYPE:            application_security_group.New(),
		data_factory.TYPE:                          data_factory.New(),
		data_factory_integration_runtime.TYPE:      data_factory_integration_runtime.New(),
		data_factory_managed_private_endpoint.TYPE: data_factory_managed_private_endpoint.New(),
		dns_record.TYPE:                            dns_record.New(),
		function_app.TYPE:                          function_app.New(),
		key_vault.TYPE:                             key_vault.New(),
		load_balancer.TYPE:                         load_balancer.New(),
		load_balancer_frontend.TYPE:                load_balancer_frontend.New(),
		//log_analytics.TYPE:                         log_analytics.New(),
		nat_gateway.TYPE:       nat_gateway.New(),
		network_interface.TYPE: network_interface.New(),
		//network_security_group.TYPE:                network_security_group.New(),
		private_dns_zone.TYPE:          private_dns_zone.New(),
		private_endpoint.TYPE:          private_endpoint.New(),
		private_link_service.TYPE:      private_link_service.New(),
		public_ip_address.TYPE:         public_ip_address.New(),
		sql_server.TYPE:                sql_server.New(),
		storage_account.TYPE:           storage_account.New(),
		subnet.TYPE:                    subnet.New(),
		virtual_machine.TYPE:           virtual_machine.New(),
		virtual_machine_scale_set.TYPE: virtual_machine_scale_set.New(),
		virtual_network.TYPE:           virtual_network.New(),
	}
	resource_map = map[string]*node.ResourceAndNode{}
)

type drawio struct {
}

func New() *drawio {
	return &drawio{}
}

func (d *drawio) WriteDiagram(filename string, resources []*az.Resource) {
	// at this point only the Azure resources are known - this function adds the corresponding DrawIO icons
	cells := addDrawIOCells(resources)

	// with every DrawIO icon present, add the dependency arrows
	dependencyArrows := addDependencyArrows()

	// some resources like vnets and subnets needs boxes draw around them, and their resources moved into them
	boxes := addBoxes()

	// combine everything and render them in the final diagram
	// items appended first are rendered first (in the background)
	cellsToRender := []string{}
	cellsToRender = append(cellsToRender, boxes...)
	cellsToRender = append(cellsToRender, dependencyArrows...)
	cellsToRender = append(cellsToRender, list.Map(cells, node.ToMXCell)...)

	if err := writeDiagram(filename, cellsToRender); err != nil {
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
		log.Printf("unhandled type %s", resource.Type)
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

	processed_nodes := map[string]*node.Node{}
	for k, v := range resource_map {
		processed_nodes[k] = v.Node
	}

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

			ok = resource_map[dependency].Node != nil

			if !ok {
				log.Printf("target %s was not drawn, skipping ...", dependency)
				continue
			}

			target := resource_map[dependency].Resource

			dependency := f.DrawDependency(resource, target, &processed_nodes)

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

	// move all resources to a starting point
	for id, resourceAndNode := range resource_map {
		resources = append(resources, resourceAndNode.Resource)

		resource_map[id].Node.SetPosition(-200, -200)
	}

	padding := 45
	boxOriginX := 0
	maxHeightSoFar := 0

	// 1. handle subnets
	subnetNodes := processSubnets(resources, &maxHeightSoFar, &boxOriginX, padding)

	// 2. handle vnets
	vnetNodes := processVnets(resources, padding, boxOriginX, maxHeightSoFar)

	subnetCells := list.Map(subnetNodes, node.ToMXCell)
	vnetCells := list.Map(vnetNodes, node.ToMXCell)

	// return vnets first so they are rendered in the background
	return append(vnetCells, subnetCells...)
}

func getResourcesInSubet(resources []*az.Resource, subnetId string) []*node.ResourceAndNode {
	azResourcesInSubnet := list.Filter(resources, func(resource *az.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == subnetId })
	})
	resourcesInSubnet := list.Map(azResourcesInSubnet, func(resource *az.Resource) *node.ResourceAndNode {
		return resource_map[resource.Id]
	})
	return resourcesInSubnet
}

// TODO: move to handler?
func processSubnets(resources []*az.Resource, maxHeightSoFar, boxOriginX *int, padding int) []*node.Node {
	nodes := []*node.Node{}

	subnetsToProcess := list.Filter(resources, func(resource *az.Resource) bool { return resource.Type == az.SUBNET })

	// ensure some deterministic order
	sort.Slice(subnetsToProcess, func(i, j int) bool {
		return subnetsToProcess[i].Name < subnetsToProcess[j].Name
	})

	for _, subnet := range subnetsToProcess {
		// 1.1 determine what resources belongs in a subnet
		resourcesInSubnet := getResourcesInSubet(resources, subnet.Id)

		// ensure some deterministic order
		sort.Slice(resourcesInSubnet, func(i, j int) bool {
			return resourcesInSubnet[i].Resource.Name < resourcesInSubnet[j].Resource.Name
		})

		// 1.2 determine the width and height of the subnet box
		subnetNode := resource_map[subnet.Id].Node
		subnetNodePosition := subnetNode.GetProperties()

		width := list.Fold(resourcesInSubnet, 0, func(r *node.ResourceAndNode, acc int) int { return acc + r.Node.GetProperties().Width })
		height := list.Fold(resourcesInSubnet, 0, func(r *node.ResourceAndNode, acc int) int {
			return int(math.Max(float64(acc), float64(r.Node.GetProperties().Height)))
		})

		height += padding
		width += (padding * len(resourcesInSubnet))

		*maxHeightSoFar = int(math.Max(float64(*maxHeightSoFar), float64(height)))

		boxProperties := &node.Properties{
			X:      *boxOriginX,
			Y:      0,
			Width:  width,
			Height: height,
		}

		// 1.3 move the subnet icon to the edge of the box
		offsetX := boxProperties.X - subnetNodePosition.Width/2
		offsetY := boxProperties.Y - subnetNodePosition.Height/2
		subnetNode.SetPosition(offsetX, offsetY)

		// 1.4 move all resources in the subnet, inside the box
		acc := boxProperties.X + padding // start of box
		for _, resource := range resourcesInSubnet {
			offsetX := acc
			offsetY := boxProperties.Height/2 - resource.Node.GetProperties().Height/2
			resource.Node.SetPosition(offsetX, offsetY)
			acc += resource.Node.GetProperties().Width + padding
		}

		boxProperties.Width += padding

		// 1.5 adjust padding between the current box and the next subnets box on the X axis
		*boxOriginX += boxProperties.Width + (subnetNodePosition.Width/2 + padding)

		subnetBox := node.NewBox(boxProperties)

		nodes = append(nodes, subnetBox)
	}

	return nodes
}

// TODO: move to handler?
func processVnets(resources []*az.Resource, padding, boxOriginX, maxHeightSoFar int) []*node.Node {
	nodes := []*node.Node{}

	vnetsToProcess := list.Filter(resources, func(resource *az.Resource) bool { return resource.Type == az.VIRTUAL_NETWORK })

	for _, vnet := range vnetsToProcess {
		vnetNode := resource_map[vnet.Id].Node

		// assuming there exists only one vnet
		// TODO: handle multiple vnets?

		// 2.1 determine the width of the box
		// first subnet is located at (0,0) so the vnet box should be created a bit to the left and above
		// TODO: don't rely on (0,0)
		boxProperties := &node.Properties{
			X:      0 - padding,
			Y:      0 - padding,
			Width:  boxOriginX + padding,
			Height: maxHeightSoFar + (2 * padding),
		}

		// 2.2 move the vnet icon to the bottom-left of the box
		offsetX := boxProperties.X - vnetNode.GetProperties().Width/2
		offsetY := boxProperties.Y + boxProperties.Height - vnetNode.GetProperties().Height/2
		vnetNode.SetPosition(offsetX, offsetY)

		vnetBox := node.NewBox(boxProperties)

		nodes = append(nodes, vnetBox)
	}

	return nodes
}

func writeDiagram(filename string, cells []string) error {
	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	var buffer bytes.Buffer

	for _, cell := range cells {
		buffer.WriteString(cell)
	}

	diagramId := guid.NewGuidAlphanumeric()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(fmt.Sprintf(`<mxfile host="Electron" agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) draw.io/25.0.1 Chrome/128.0.6613.186 Electron/32.2.6 Safari/537.36" version="25.0.1">
	<diagram name="Page-1" id="%s">
		<mxGraphModel dx="2074" dy="1196" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="850" pageHeight="1100" math="0" shadow="0">
			<root>
			<mxCell id="0" />
			<mxCell id="1" parent="0" />
			%s
			</root>
		</mxGraphModel>
	</diagram>
</mxfile>
`, diagramId, buffer.String()))

	if err != nil {
		return err
	}

	return w.Flush()
}
