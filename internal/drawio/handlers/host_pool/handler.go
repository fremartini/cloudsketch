package host_pool

import (
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.HOST_POOL
	IMAGE  = images.HOST_POOL
	WIDTH  = 68
	HEIGHT = 68
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
	typeBlacklist := []string{types.SUBSCRIPTION}

	targetResources := list.Map(targets, func(target *models.Resource) *node.ResourceAndNode {
		return (*resource_map)[target.Id]
	})

	targetResources = list.Filter(targetResources, func(target *node.ResourceAndNode) bool {
		return !list.Contains(typeBlacklist, func(t string) bool {
			return target.Resource.Type == t
		})
	})

	sourceNode := (*resource_map)[source.Id].Node

	arrows := list.Fold(targetResources, []*node.Arrow{}, func(target *node.ResourceAndNode, acc []*node.Arrow) []*node.Arrow {
		return append(acc, node.NewArrow(sourceNode.Id(), target.Node.Id(), nil))
	})

	return arrows
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
