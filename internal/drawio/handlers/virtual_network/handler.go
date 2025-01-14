package virtual_network

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/diagram"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
)

type handler struct{}

const (
	TYPE   = az.VIRTUAL_NETWORK
	IMAGE  = images.VIRTUAL_NETWORK
	WIDTH  = 67
	HEIGHT = 40
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	vnet := node.NewIcon(IMAGE, resource.Name, &node.Properties{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	})

	return []*node.Node{vnet}
}

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	properties := &node.Properties{
		X:      0,
		Y:      0,
		Width:  diagram.BoxOriginX,
		Height: diagram.MaxHeightSoFar,
	}

	vnetsToProcess := list.Filter(resources, func(resource *az.Resource) bool { return resource.Type == az.VIRTUAL_NETWORK })

	for _, vnet := range vnetsToProcess {
		vnetNode := (*resource_map)[vnet.Id].Node
		vnetNodeProperties := vnetNode.GetProperties()

		// assuming there exists only one vnet
		// TODO: handle multiple vnets?

		// move the box a bit to the left and above to fit its children
		properties = &node.Properties{
			X:      properties.X - diagram.Padding,
			Y:      properties.Y - diagram.Padding,
			Width:  properties.Width + diagram.Padding,
			Height: properties.Height + (2 * diagram.Padding),
		}

		// move the vnet icon to the bottom-left of the box
		offsetX := properties.X - vnetNodeProperties.Width/2
		offsetY := properties.Y + properties.Height - vnetNodeProperties.Height/2
		vnetNode.SetPosition(offsetX, offsetY)

		vnetBox := node.NewBox(properties)
		nodes = append(nodes, vnetBox)
	}

	return nodes
}
