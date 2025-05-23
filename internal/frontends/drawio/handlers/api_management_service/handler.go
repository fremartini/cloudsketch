package api_management_service

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
)

type handler struct{}

const (
	TYPE   = types.API_MANAGEMENT_SERVICE
	IMAGE  = images.API_MANAGEMENT_SERVICE
	WIDTH  = 65
	HEIGHT = 60
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

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{})
}

func (*handler) GroupResources(resource *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInAPIM := node.GetChildResourcesOfType(resources, resource.Id, types.API_MANAGEMENT_API, resource_map)

	if len(resourcesInAPIM) == 0 {
		return []*node.Node{}
	}

	apimNode := (*resource_map)[resource.Id].Node

	box := node.BoxResources(apimNode, resourcesInAPIM)

	return []*node.Node{box}
}
