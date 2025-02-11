package dns_record

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
)

type handler struct{}

const (
	TYPE   = az.DNS_RECORD
	WIDTH  = 45
	HEIGHT = 45
)

func New() *handler {
	return &handler{}
}

func (*handler) MapResource(resource *az.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	return node.NewGeneric(map[string]interface{}{
		"style": "shadow=0;dashed=0;html=1;strokeColor=none;fillColor=#4495D1;labelPosition=center;verticalLabelPosition=bottom;verticalAlign=top;align=center;outlineConnect=0;shape=mxgraph.veeam.dns;",
		"value": resource.Name,
	}, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	sourceNode := (*resource_map)[source.Id].Node
	targetNode := (*resource_map)[target.Id].Node

	// dns records can be contained inside aprivate DNS Zone. Don't draw these
	if sourceNode.ContainedIn == targetNode.ContainedIn {
		return nil
	}

	return node.NewArrow(sourceNode.Id(), targetNode.Id())
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
