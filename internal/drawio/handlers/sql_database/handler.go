package sql_database

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.SQL_DATABASE
	IMAGE  = images.SQL_DATABASE
	WIDTH  = 48
	HEIGHT = 64
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(_ *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
