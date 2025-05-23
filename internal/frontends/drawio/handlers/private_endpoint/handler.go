package private_endpoint

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
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
	sourceSubnet := getSubnet(source)

	attachedToId := source.Properties["attachedTo"]

	attachedTo := (*resource_map)[attachedToId[0]]

	attachedToSubnet := getSubnet(attachedTo.Resource)

	sourceNode := (*resource_map)[source.Id]

	if attachedToSubnet != nil && sourceSubnet != attachedToSubnet {
		return []*node.Arrow{node.NewArrow(sourceNode.Node.Id(), attachedTo.Node.Id(), nil)}
	}

	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{types.SUBNET})
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
