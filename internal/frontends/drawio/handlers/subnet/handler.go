package subnet

import (
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/frontends/drawio/handlers/diagram"
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
	"fmt"
)

type handler struct{}

const (
	TYPE   = types.SUBNET
	IMAGE  = images.SUBNET
	WIDTH  = 68
	HEIGHT = 41
)

var (
	STYLE = "fillColor=#7EA6E0;strokeColor=#6c8ebf;opacity=50;"
)

func New() *handler {
	return &handler{}
}

func (*handler) MapResource(resource *models.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	subnetSize := resource.Properties["size"][0]

	name := fmt.Sprintf("%s/%s", resource.Name, subnetSize)

	link := resource.GetLinkOrDefault()

	return node.NewIcon(IMAGE, name, &geometry, link)
}

func getResourcseOfType(resource *models.Resource, resource_map *map[string]*node.ResourceAndNode, typ string) []*models.Resource {
	return list.Filter(resource.DependsOn, func(dependency *models.Resource) bool {
		r, ok := (*resource_map)[dependency.Id]

		if !ok {
			return false
		}

		return r.Resource.Type == typ
	})
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	var parentGroup *node.Node = nil

	routeTables := getResourcseOfType(resource.Resource, resource_map, types.ROUTE_TABLE)
	if len(routeTables) == 1 {
		routeTable := (*resource_map)[routeTables[0].Id]

		parentGroup = node.GroupIconsAndSetPosition(resource.Node, routeTable.Node, node.TOP_LEFT)
	}

	networkSecurityGroups := getResourcseOfType(resource.Resource, resource_map, types.NETWORK_SECURITY_GROUP)

	if len(networkSecurityGroups) == 1 {
		networkSecurityGroup := (*resource_map)[networkSecurityGroups[0].Id]

		// other subnets might point to the same NSG. If they do, ignore the merging
		if snets := resourcesWithReferencesTo(resource_map, networkSecurityGroup.Resource.Id); snets != 1 {
			return parentGroup
		}

		if parentGroup == nil {
			// route table icon was not set
			return node.GroupIconsAndSetPosition(resource.Node, networkSecurityGroup.Node, node.TOP_RIGHT)
		}

		// route table icon was set
		networkSecurityGroupGeometry := networkSecurityGroup.Node.GetGeometry()

		networkSecurityGroup.Node.SetProperty("parent", parentGroup.Id())
		networkSecurityGroup.Node.SetDimensions(networkSecurityGroupGeometry.Width/2, networkSecurityGroupGeometry.Width/2)

		node.SetIconRelativeTo(networkSecurityGroup.Node, resource.Node.GetParentOrThis(), node.TOP_RIGHT)
		networkSecurityGroup.Node.ContainedIn = parentGroup
		networkSecurityGroup.Node.SetProperty("value", "")
	}

	return parentGroup
}

func resourcesWithReferencesTo(resource_map *map[string]*node.ResourceAndNode, resourceId string) int {
	count := 0

	for _, v := range *resource_map {
		if list.Contains(v.Resource.DependsOn, func(d *models.Resource) bool {
			return d.Id == resourceId
		}) {
			count++
		}
	}

	return count
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{types.VIRTUAL_NETWORK})
}

func (*handler) GroupResources(subnet *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInSubnet := node.GetChildResources(resources, subnet.Id, resource_map)

	if len(resourcesInSubnet) == 0 {
		return nil
	}

	nodes := list.Map(resourcesInSubnet, func(ran *node.ResourceAndNode) *node.Node {
		return ran.Node.GetParentOrThis()
	})

	// a subnet can contain resources that belong to the same group, these needs to be filtered to
	// avoid moving the same group multiple times
	seenGroups := set.New[string]()

	nodes = list.Filter(nodes, func(n *node.Node) bool {
		if seenGroups.Contains(n.Id()) {
			return false
		}

		seenGroups.Add(n.Id())

		return true
	})

	subnetNode := (*resource_map)[subnet.Id].Node

	// subnets can be in a group because of UDRs
	subnetNode = subnetNode.GetParentOrThis()

	box := node.NewBox(&node.Geometry{
		X:      0,
		Y:      0,
		Width:  0,
		Height: 0,
	}, &STYLE)

	subnetNode.SetProperty("parent", box.Id())
	subnetNode.ContainedIn = box
	node.SetIconRelativeTo(subnetNode, box, node.TOP_LEFT)

	node.FillResourcesInBox(box, nodes, diagram.Padding, true)

	return []*node.Node{box}
}
