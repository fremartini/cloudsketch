package virtual_network

import (
	"cloudsketch/internal/drawio/handlers/diagram"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
	"fmt"
)

type handler struct{}

const (
	TYPE   = types.VIRTUAL_NETWORK
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

func (*handler) MapResource(resource *models.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	vnetSize, ok := resource.Properties["size"]

	if !ok {
		return node.NewIcon(IMAGE, resource.Name, &geometry)
	}

	name := fmt.Sprintf("%s/%s", resource.Name, vnetSize)

	return node.NewIcon(IMAGE, name, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return []*node.Arrow{}
}

func (*handler) GroupResources(vnet *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
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
