package web_sites

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"log"
)

type handler struct{}

const (
	TYPE   = az.WEB_SITES
	WIDTH  = 68
	HEIGHT = 60
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource) *node.Node {
	var image = ""

	subtype := resource.Properties["subType"]

	switch subtype {
	case az.APP_SERVICE_SUBTYPE:
		image = images.APP_SERVICE
	case az.FUNCTION_APP_SUBTYPE:
		image = images.FUNCTION_APP
	default:
		log.Fatalf("No image registered for subtype %s", subtype)
	}

	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
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

	// function apps can be contained inside an app service plan. Don't draw these
	if sourceNode.ContainedIn == targetNode.ContainedIn {
		return nil
	}

	return node.NewArrow(sourceNode.Id(), targetNode.Id())
}

func (*handler) DrawBox(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
