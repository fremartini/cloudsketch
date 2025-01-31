package virtual_network

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/diagram"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.VIRTUAL_NETWORK
	IMAGE  = images.VIRTUAL_NETWORK
	WIDTH  = 67
	HEIGHT = 40
)

var (
	STYLE = "fillColor=#dae8fc;strokeColor=#6c8ebf"
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	vnet := node.NewIcon(IMAGE, resource.Name, &node.Geometry{
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

func (*handler) DrawBox(vnet *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	geometry := &node.Geometry{
		X:      0,
		Y:      0,
		Width:  diagram.BoxOriginX,
		Height: diagram.MaxHeightSoFar,
	}

	vnetNode := (*resource_map)[vnet.Id].Node
	vnetNodegeometry := vnetNode.GetGeometry()

	// assuming there exists only one vnet
	// TODO: handle multiple vnets?

	// move the box a bit to the left and above to fit its children
	geometry = &node.Geometry{
		X:      geometry.X - diagram.Padding,
		Y:      geometry.Y - diagram.Padding,
		Width:  geometry.Width + diagram.Padding,
		Height: geometry.Height + (2 * diagram.Padding),
	}

	// move the vnet icon to the bottom-left of the box
	offsetX := geometry.X - vnetNodegeometry.Width/2
	offsetY := geometry.Y + geometry.Height - vnetNodegeometry.Height/2
	vnetNode.SetPosition(offsetX, offsetY)

	vnetBox := node.NewBox(geometry, &STYLE)
	nodes = append(nodes, vnetBox)

	return nodes
}
