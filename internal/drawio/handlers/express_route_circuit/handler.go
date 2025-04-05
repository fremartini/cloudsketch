package express_route_circuit

import (
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.EXPRESS_ROUTE_CIRCUIT
	IMAGE  = images.EXPRESS_ROUTE_CIRCUIT
	WIDTH  = 70
	HEIGHT = 64
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
	arrows := node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{})

	peerings, ok := source.Properties["peerings"]

	if !ok {
		return arrows
	}

	arrows = list.Fold(peerings, arrows, func(peering string, acc []*node.Arrow) []*node.Arrow {
		return addDependencyToPeering(peering, source, resource_map)
	})

	return arrows
}

func addDependencyToPeering(peering string, source *models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	expressRouteGatewaysWithPeering := []*node.ResourceAndNode{}
	for _, r := range *resource_map {
		expressRouteGatewaysWithPeering = append(expressRouteGatewaysWithPeering, r)
	}

	expressRouteGatewaysWithPeering = list.Filter(expressRouteGatewaysWithPeering, func(ran *node.ResourceAndNode) bool {
		if ran.Resource.Type != types.EXPRESS_ROUTE_GATEWAY {
			return false
		}

		gatewayPeerings, ok := ran.Resource.Properties["peerings"]

		if !ok {
			return false
		}

		return list.Contains(gatewayPeerings, func(gatewayPeering string) bool {
			return gatewayPeering == peering
		})
	})

	sourceNode := (*resource_map)[source.Id].Node

	arrows := list.Map(expressRouteGatewaysWithPeering, func(peering *node.ResourceAndNode) *node.Arrow {
		return node.NewArrow(sourceNode.Id(), peering.Node.Id(), nil)
	})

	return arrows
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
