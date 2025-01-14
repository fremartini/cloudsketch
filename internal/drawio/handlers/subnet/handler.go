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

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	subnet := node.NewIcon(IMAGE, resource.Name, &node.Properties{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	})

	return []*node.Node{subnet}
}

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
	// don't draw arrows to virtual networks
	if target.Type == az.VIRTUAL_NETWORK {
		return nil
	}

	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	subnetsToProcess := list.Filter(resources, func(resource *az.Resource) bool { return resource.Type == az.SUBNET })

	// ensure some deterministic order
	sort.Slice(subnetsToProcess, func(i, j int) bool {
		return subnetsToProcess[i].Name < subnetsToProcess[j].Name
	})

	for _, subnet := range subnetsToProcess {
		// determine what resources belongs in a subnet
		resourcesInSubnet := getResourcesInSubet(resources, subnet.Id, resource_map)

		// ensure some deterministic order
		sort.Slice(resourcesInSubnet, func(i, j int) bool {
			return resourcesInSubnet[i].Resource.Name < resourcesInSubnet[j].Resource.Name
		})

		// determine the width and height of the subnet box
		subnetNode := (*resource_map)[subnet.Id].Node
		subnetNodePosition := subnetNode.GetProperties()

		width := list.Fold(resourcesInSubnet, 0, func(r *node.ResourceAndNode, acc int) int { return acc + r.Node.GetProperties().Width })
		height := list.Fold(resourcesInSubnet, 0, func(r *node.ResourceAndNode, acc int) int {
			return int(math.Max(float64(acc), float64(r.Node.GetProperties().Height)))
		})

		height += diagram.Padding
		width += (diagram.Padding * len(resourcesInSubnet))

		diagram.MaxHeightSoFar = int(math.Max(float64(diagram.MaxHeightSoFar), float64(height)))

		boxProperties := &node.Properties{
			X:      diagram.BoxOriginX,
			Y:      0,
			Width:  width,
			Height: height,
		}

		// move the subnet icon to the edge of the box
		offsetX := boxProperties.X - subnetNodePosition.Width/2
		offsetY := boxProperties.Y - subnetNodePosition.Height/2
		subnetNode.SetPosition(offsetX, offsetY)

		// move all resources in the subnet, inside the box
		acc := boxProperties.X + diagram.Padding // start of box
		for _, resource := range resourcesInSubnet {
			offsetX := acc
			offsetY := boxProperties.Height/2 - resource.Node.GetProperties().Height/2
			resource.Node.SetPosition(offsetX, offsetY)
			acc += resource.Node.GetProperties().Width + diagram.Padding
		}

		boxProperties.Width += diagram.Padding

		// adjust padding between the current box and the next subnets box on the X axis
		diagram.BoxOriginX += boxProperties.Width + (subnetNodePosition.Width/2 + diagram.Padding)

		subnetBox := node.NewBox(boxProperties)

		nodes = append(nodes, subnetBox)
	}

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
