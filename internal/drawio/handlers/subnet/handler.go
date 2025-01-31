package subnet

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/diagram"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
	"math"
	"sort"
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
	nodes := []*node.Node{}

	// determine what resources belongs in a subnet
	resourcesInSubnet := getResourcesInSubet(resources, subnet.Id, resource_map)

	// ensure some deterministic order
	sort.Slice(resourcesInSubnet, func(i, j int) bool {
		return resourcesInSubnet[i].Resource.Name < resourcesInSubnet[j].Resource.Name
	})

	// determine the width and height of the subnet box
	subnetNode := (*resource_map)[subnet.Id].Node
	subnetNodePosition := subnetNode.GetGeometry()

	width := list.Fold(resourcesInSubnet, 0, func(r *node.ResourceAndNode, acc int) int {
		if r.Node.ContainedIn != nil {
			return acc
		}

		return acc + r.Node.GetGeometry().Width
	})
	height := list.Fold(resourcesInSubnet, 0, func(r *node.ResourceAndNode, acc int) int {
		if r.Node.ContainedIn != nil {
			return int(math.Max(float64(acc), float64(r.Node.ContainedIn.GetGeometry().Height)))
		}

		return int(math.Max(float64(acc), float64(r.Node.GetGeometry().Height)))
	})

	height += diagram.Padding
	width += (diagram.Padding * len(resourcesInSubnet))

	diagram.MaxHeightSoFar = int(math.Max(float64(diagram.MaxHeightSoFar), float64(height)))

	boxgeometry := &node.Geometry{
		X:      diagram.BoxOriginX,
		Y:      0,
		Width:  width,
		Height: height,
	}

	// move the subnet icon to the edge of the box
	offsetX := boxgeometry.X - subnetNodePosition.Width/2
	offsetY := boxgeometry.Y - subnetNodePosition.Height/2
	subnetNode.SetPosition(offsetX, offsetY)

	box := node.NewBox(boxgeometry, &STYLE)

	// TODO: can this be refacted to use FillResourcesInBoxLinear?

	// move all resources in the subnet, inside the box
	acc := boxgeometry.X + diagram.Padding // start of box
	movedBoxes := map[string]bool{}

	for _, resource := range resourcesInSubnet {

		// If the resource is contained inside a box, the box should be moved instead of this resource
		if resource.Node.ContainedIn != nil {
			_, ok := movedBoxes[resource.Node.ContainedIn.Id()]

			// box has already been moved, skip
			if ok {
				continue
			}

			offsetX := acc
			offsetY := boxgeometry.Height/2 - resource.Node.ContainedIn.GetGeometry().Height/2
			resource.Node.ContainedIn.SetPosition(offsetX, offsetY)
			acc += resource.Node.ContainedIn.GetGeometry().Width + diagram.Padding

			movedBoxes[resource.Node.ContainedIn.Id()] = true

			continue
		}

		offsetX := acc
		offsetY := boxgeometry.Height/2 - resource.Node.GetGeometry().Height/2
		resource.Node.SetPosition(offsetX, offsetY)
		acc += resource.Node.GetGeometry().Width + diagram.Padding
	}

	boxgeometry.Width += diagram.Padding

	// adjust padding between the current box and the next subnets box on the X axis
	diagram.BoxOriginX += boxgeometry.Width + (subnetNodePosition.Width/2 + diagram.Padding)

	nodes = append(nodes, box)

	return nodes
}

func getResourcesInSubet(resources []*az.Resource, subnetId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInSubnet := list.Filter(resources, func(resource *az.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == subnetId })
	})
	resourcesInSubnet := list.Map(azResourcesInSubnet, func(resource *az.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInSubnet
}
