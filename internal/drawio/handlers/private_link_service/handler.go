package private_link_service

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.PRIVATE_LINK_SERVICE
	IMAGE  = images.PRIVATE_LINK_SERVICE
	WIDTH  = 69
	HEIGHT = 40
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	properties := node.Properties{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	n := node.NewIcon(IMAGE, resource.Name, &properties)

	return []*node.Node{n}
}

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}
