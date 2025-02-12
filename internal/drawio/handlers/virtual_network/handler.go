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

func (*handler) MapResource(resource *az.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	return node.NewIcon(IMAGE, resource.Name, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source *az.Resource, targets []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return []*node.Arrow{}
}

func (*handler) GroupResources(vnet *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	geometry := &node.Geometry{
		X:      0,
		Y:      0,
		Width:  diagram.BoxOriginX,
		Height: diagram.MaxHeightSoFar,
	}

	vnetNode := (*resource_map)[vnet.Id].Node
	vnetNodegeometry := vnetNode.GetGeometry()

	// move the box a bit to the left and above to fit its children
	geometry = &node.Geometry{
		X:      geometry.X - diagram.Padding,
		Y:      geometry.Y - diagram.Padding,
		Width:  geometry.Width + diagram.Padding,
		Height: geometry.Height + (2 * diagram.Padding),
	}

	box := node.NewBox(geometry, &STYLE)

	vnetNode.SetProperty("parent", box.Id())
	vnetNode.SetPosition(-vnetNodegeometry.Width/2, geometry.Height-vnetNodegeometry.Height/2)

	return []*node.Node{box}
}
