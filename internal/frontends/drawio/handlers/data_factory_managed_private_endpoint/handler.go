package data_factory_managed_private_endpoint

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/handlers/private_endpoint"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
)

type handler struct{}

const (
	TYPE   = types.DATA_FACTORY_MANAGED_PRIVATE_ENDPOINT
	IMAGE  = images.DATA_FACTORY_MANAGED_PRIVATE_ENDPOINT
	WIDTH  = private_endpoint.WIDTH
	HEIGHT = private_endpoint.HEIGHT
)

func New() *handler {
	return &handler{}
}

func (*handler) MapResource(resource *models.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH / 2,
		Height: HEIGHT / 2,
	}

	link := resource.GetLinkOrDefault()

	return node.NewIcon(IMAGE, resource.Name, &geometry, link)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTargets(source, targets, resource_map, []string{types.DATA_FACTORY})
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
