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

	// max height is required to determine the center on the Y axis to place icons in the middle
	height := list.Fold(resourcesInSubnet, 0, func(r *node.ResourceAndNode, acc int) int {
		if r.Node.ContainedIn != nil {
			return int(math.Max(float64(acc), float64(r.Node.ContainedIn.GetGeometry().Height)))
		}

		return int(math.Max(float64(acc), float64(r.Node.GetGeometry().Height)))
	})

	height += diagram.Padding

	geometry := &node.Geometry{
		X:      diagram.BoxOriginX,
		Y:      0,
		Width:  0,
		Height: height,
	}

	// move the subnet icon to the edge of the box
	subnetNode := (*resource_map)[subnet.Id].Node
	subnetNodePosition := subnetNode.GetGeometry()
	subnetNode.SetPosition(geometry.X-subnetNodePosition.Width/2, geometry.Y-subnetNodePosition.Height/2)

	box := node.NewBox(geometry, &STYLE)

	// move all resources in the subnet, inside the box
	acc := geometry.X + diagram.Padding // start of box
	movedGroups := map[string]bool{}

	for _, resource := range resourcesInSubnet {
		// if the resource is contained inside a box, the box should be moved instead of this resource but only if it has not been moved
		if resource.Node.ContainedIn != nil {
			_, ok := movedGroups[resource.Node.ContainedIn.Id()]

			// box has already been moved
			if ok {
				continue
			}

			offsetX := acc
			offsetY := geometry.Height/2 - resource.Node.ContainedIn.GetGeometry().Height/2
			resource.Node.ContainedIn.SetPosition(offsetX, offsetY)
			acc += resource.Node.ContainedIn.GetGeometry().Width + diagram.Padding

			geometry.Width += resource.Node.ContainedIn.GetGeometry().Width + diagram.Padding

			movedGroups[resource.Node.ContainedIn.Id()] = true

			continue
		}

		offsetX := acc
		offsetY := geometry.Height/2 - resource.Node.GetGeometry().Height/2
		resource.Node.SetPosition(offsetX, offsetY)
		acc += resource.Node.GetGeometry().Width + diagram.Padding

		geometry.Width += resource.Node.GetGeometry().Width + diagram.Padding
	}

	geometry.Width += diagram.Padding

	// adjust padding between the current box and the next subnets box on the X axis
	diagram.BoxOriginX += geometry.Width + (subnetNodePosition.Width/2 + diagram.Padding)

	// the vnet needs to know about the tallest vnet so it can fit it
	diagram.MaxHeightSoFar = int(math.Max(float64(diagram.MaxHeightSoFar), float64(height)))

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
