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

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	subnet := node.NewIcon(IMAGE, resource.Name, &node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	})

	return []*node.Node{subnet}
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	// don't draw arrows to virtual networks
	if target.Type == az.VIRTUAL_NETWORK {
		return nil
	}

	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(subnet *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInSubnet := getResourcesInSubnet(resources, subnet.Id, resource_map)

	seenGroups := map[string]bool{}

	// resources in the subnet can belong to the same group - only count the group once
	resourcesInSubnet = list.Filter(resourcesInSubnet, func(r *node.ResourceAndNode) bool {
		if r.Node.ContainedIn == nil {
			return true
		}

		if _, ok := seenGroups[r.Node.ContainedIn.Id()]; ok {
			return false
		}

		seenGroups[r.Node.ContainedIn.Id()] = true

		return true
	})

	geometry := &node.Geometry{
		X:      diagram.BoxOriginX,
		Y:      0,
		Width:  0,
		Height: 0,
	}

	// move the subnet icon to the edge of the box
	subnetNode := (*resource_map)[subnet.Id].Node
	subnetNodePosition := subnetNode.GetGeometry()
	subnetNode.SetPosition(geometry.X-subnetNodePosition.Width/2, geometry.Y-subnetNodePosition.Height/2)

	box := node.NewBox(geometry, &STYLE)

	node.FillResourcesInBox(box, resourcesInSubnet, diagram.Padding)

	// adjust padding between the current box and the next subnets box on the X axis
	diagram.BoxOriginX = geometry.X + geometry.Width + (subnetNodePosition.Width/2 + diagram.Padding)

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
