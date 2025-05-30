package bastion

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.BASTION
	IMAGE  = images.BASTION
	WIDTH  = 58
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
	publicIps := list.Filter(resource.Resource.DependsOn, func(dependency *models.Resource) bool {
		r, ok := (*resource_map)[dependency.Id]

		if !ok {
			return false
		}

		return r.Resource.Type == types.PUBLIC_IP_ADDRESS
	})

	if len(publicIps) == 1 {
		pipResource := (*resource_map)[publicIps[0].Id]
		return node.GroupIconsAndSetPosition(resource.Node, pipResource.Node, node.TOP_RIGHT)
	}

	return nil
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTargets(source, targets, resource_map, []string{types.SUBNET, types.PUBLIC_IP_ADDRESS})
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
