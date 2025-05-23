package container_apps_environment

import (
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.CONTAINER_APPS_ENVIRONMENT
	IMAGE  = images.CONTAINER_APPS_ENVIRONMENT
	WIDTH  = 68
	HEIGHT = 68
)

var (
	STYLE = "rounded=0;whiteSpace=wrap;html=1;dashed=1;opacity=50;"
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

func (*handler) GroupResources(containerEnvironment *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInContainerEnvironment := node.GetChildResourcesOfType(resources, containerEnvironment.Id, types.CONTAINER_APP, resource_map)

	if len(resourcesInContainerEnvironment) == 0 {
		return []*node.Node{}
	}

	seenGroups := set.New[string]()

	resourcesInContainerEnvironment = list.Filter(resourcesInContainerEnvironment, func(r *node.ResourceAndNode) bool {
		n := r.Node.GetParentOrThis()

		if seenGroups.Contains(n.Id()) {
			return false
		}

		seenGroups.Add(n.Id())

		return true
	})

	containerEnvironmentNode := (*resource_map)[containerEnvironment.Id].Node

	box := node.BoxResources(containerEnvironmentNode, resourcesInContainerEnvironment)

	return []*node.Node{box}
}
