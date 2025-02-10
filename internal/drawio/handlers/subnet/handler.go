package subnet

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/diagram"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
	"math"
)

type handler struct{}

const (
	TYPE   = az.SUBNET
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

func (*handler) DrawIcon(resource *az.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	return node.NewIcon(IMAGE, resource.Name, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	routeTables := list.Filter(resource.Resource.DependsOn, func(dependency string) bool {
		r, ok := (*resource_map)[dependency]

		if !ok {
			return false
		}

		return r.Resource.Type == az.ROUTE_TBLE
	})

	if len(routeTables) != 1 {
		return nil
	}

	routeTable := (*resource_map)[routeTables[0]]

	return node.SetIcon(resource.Node, routeTable.Node, node.TOP_LEFT)
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	// don't draw arrows to virtual networks
	if target.Type == az.VIRTUAL_NETWORK {
		return nil
	}

	sourceNode := (*resource_map)[source.Id].Node
	targetNode := (*resource_map)[target.Id].Node

	// if they are in the same group, don't draw the arrow
	if sourceNode.ContainedIn != nil && targetNode.ContainedIn != nil {
		if sourceNode.ContainedIn == targetNode.ContainedIn {
			return nil
		}
	}

	return node.NewArrow(sourceNode.Id(), targetNode.Id())
}

func (*handler) DrawBox(subnet *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInSubnet := getResourcesInSubnet(resources, subnet.Id, resource_map)

	// a subnet can contain resources that belong to the same group, these needs to be filtered to
	// avoid moving the same group multiple times
	seenGroups := map[string]bool{}

	resourcesInSubnet = list.Filter(resourcesInSubnet, func(r *node.ResourceAndNode) bool {
		n := r.Node.GetParentOrThis()

		if seenGroups[n.Id()] {
			return false
		}

		seenGroups[n.Id()] = true

		return true
	})

	geometry := &node.Geometry{
		X:      diagram.BoxOriginX,
		Y:      0,
		Width:  0,
		Height: 0,
	}

	box := node.NewBox(geometry, &STYLE)

	subnetNode := (*resource_map)[subnet.Id].Node
	subnetNodeGeometry := subnetNode.GetGeometry()

	subnetNode.SetProperty("parent", box.Id())
	subnetNode.SetPosition(-subnetNodeGeometry.Width/2, -subnetNodeGeometry.Height/2)

	nodesToMove := list.Map(resourcesInSubnet, func(r *node.ResourceAndNode) *node.Node {
		return r.Node.GetParentOrThis()
	})

	node.FillResourcesInBox(box, nodesToMove, diagram.Padding)

	// adjust padding between the current box and the next subnets box on the X axis
	diagram.BoxOriginX = geometry.X + geometry.Width + (subnetNodeGeometry.Width/2 + diagram.Padding)

	// the vnet needs to know about the tallest vnet so it can fit it
	diagram.MaxHeightSoFar = int(math.Max(float64(diagram.MaxHeightSoFar), float64(box.GetGeometry().Height)))

	return []*node.Node{box}
}

func getResourcesInSubnet(resources []*az.Resource, subnetId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInSubnet := list.Filter(resources, func(resource *az.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == subnetId })
	})
	resourcesInSubnet := list.Map(azResourcesInSubnet, func(resource *az.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInSubnet
}
