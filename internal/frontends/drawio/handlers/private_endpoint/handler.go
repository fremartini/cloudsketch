package private_endpoint

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.PRIVATE_ENDPOINT
	IMAGE  = images.PRIVATE_ENDPOINT
	WIDTH  = 68
	HEIGHT = 65
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
	// don't draw arrows to the resource this private endpoint is attached to unless they are in different subnets
	attachedToIds, ok := source.Properties["attachedTo"]

	if ok {
		sourceSubnet := getSubnet(source)
		attachedToId := attachedToIds[0]

		targets = list.Filter(targets, func(dependency *models.Resource) bool {
			if dependency.Id != attachedToId {
				return true
			}

			attachedTo := (*resource_map)[attachedToId]

			attachedToSubnet := getSubnet(attachedTo.Resource)

			return attachedToSubnet != nil && sourceSubnet != attachedToSubnet
		})
	}

	return node.DrawDependencyArrowsToTargets(source, targets, resource_map, []string{types.SUBNET})
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}

func getSubnet(resource *models.Resource) *models.Resource {
	for _, dependency := range resource.DependsOn {
		if dependency.Type != types.SUBNET {
			continue
		}

		return dependency
	}

	return nil
}
