package dns_record

import (
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.DNS_RECORD
	WIDTH  = 45
	HEIGHT = 45
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

	return node.NewGeneric(map[string]any{
		"style": "shadow=0;dashed=0;html=1;strokeColor=none;fillColor=#4495D1;labelPosition=center;verticalLabelPosition=bottom;verticalAlign=top;align=center;outlineConnect=0;shape=mxgraph.veeam.dns;",
		"value": resource.Name,
	}, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	typeBlacklist := []string{types.SUBSCRIPTION, types.PRIVATE_DNS_ZONE}

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
		targetNode := target.Node

		if target.Resource.Type == types.PRIVATE_ENDPOINT {
			targetNode = targetNode.ContainedIn
		}

		return append(acc, node.NewArrow(sourceNode.Id(), targetNode.Id(), nil))
	})

	return arrows
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
