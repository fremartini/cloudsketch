package postgres_sql_server

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/handlers/sql_server"
	"cloudsketch/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.POSTGRES_SQL_SERVER
	IMAGE  = images.POSTGRES_SQL_SERVER
	WIDTH  = sql_server.WIDTH
	HEIGHT = sql_server.HEIGHT
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
	arrows := []*node.Arrow{}

	sourceId := (*resource_map)[source.Id].Node.Id()

	for _, target := range targets {
		targetId := (*resource_map)[target.Id].Node.Id()

		arrows = append(arrows, node.NewArrow(sourceId, targetId, nil))
	}

	return arrows
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
