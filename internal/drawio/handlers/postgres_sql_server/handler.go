package postgres_sql_server

import (
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/handlers/sql_server"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
)

type handler struct{}

const (
	TYPE   = types.POSTGRES_SQL_SERVER
	IMAGE  = images.POSTGRES_SQL_SERVER
	WIDTH  = sql_server.WIDTH
	HEIGHT = sql_server.HEIGHT
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

	link := resource.GetLinkOrDefault()

	return node.NewIcon(IMAGE, resource.Name, &geometry, link)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	arrows := []*node.Arrow{}

	sourceId := (*resource_map)[source.Id].Node.Id()

	for _, target := range targets {
		targetId := (*resource_map)[target.Id].Node.Id()

		arrows = append(arrows, node.NewArrow(sourceId, targetId, nil))
	}

	return arrows
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
