package web_sites

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"log"
)

type handler struct{}

const (
	TYPE = az.WEB_SITES
)

func New() *handler {
	return &handler{}
}

func (*handler) MapResource(resource *az.Resource) *node.Node {
	width := 0
	height := 0
	var image = ""

	subtype := resource.Properties["subType"]

	switch subtype {
	case az.APP_SERVICE_SUBTYPE:
		width = 68
		height = 68
		image = images.APP_SERVICE
	case az.FUNCTION_APP_SUBTYPE:
		width = 68
		height = 60
		image = images.FUNCTION_APP
	case az.LOGIC_APP_SUBTYPE:
		height = 52
		width = 67
		image = images.LOGIC_APP
	default:
		log.Fatalf("No image registered for subtype %s", subtype)
	}

	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
	}

	return node.NewIcon(image, resource.Name, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	// don't draw arrows to subnets
	if target.Type == az.SUBNET {
		return nil
	}
	sourceNode := (*resource_map)[source.Id].Node
	targetNode := (*resource_map)[target.Id].Node

	// if they are in the same group, don't draw the arrow
	if sourceNode.ContainedIn != nil && targetNode.ContainedIn != nil {
		if sourceNode.GetParentOrThis() == targetNode.GetParentOrThis() {
			return nil
		}
	}

	return node.NewArrow(sourceNode.Id(), targetNode.Id())
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
