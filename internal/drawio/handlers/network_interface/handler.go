package network_interface

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
)

type handler struct{}

const (
	TYPE   = az.NETWORK_INTERFACE
	IMAGE  = images.NETWORK_INTERFACE
	WIDTH  = 68
	HEIGHT = 60
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, resources *map[string]*node.ResourceAndNode) []*node.Node {
	for _, dependencyId := range resource.DependsOn {
		dependency := (*resources)[dependencyId]

		// TODO: why can this be the case?
		if dependency == nil {
			return []*node.Node{}
		}

		if dependency.Resource.Type == az.VIRTUAL_MACHINE || dependency.Resource.Type == az.PRIVATE_LINK_SERVICE {
			linkedResource := (*resources)[resource.Properties["attachedTo"]]
			return node.SetTopRightIcon(linkedResource, resources, IMAGE, HEIGHT, WIDTH)
		}

		// dont render NICs if they are attached to a blacklisted resource
		if isBlacklistedResource(dependency.Resource.Type) {
			return []*node.Node{}
		}
	}

	properties := node.Properties{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	n := node.NewIcon(IMAGE, resource.Name, &properties)

	return []*node.Node{n}
}

func isBlacklistedResource(resourceType string) bool {
	blacklist := []string{az.PRIVATE_ENDPOINT}

	return list.Contains(blacklist, func(e string) bool {
		return resourceType == e
	})
}

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
	// don't draw arrows to subnets
	if target.Type == az.SUBNET {
		return nil
	}

	// expect additional information on the NIC Azure resource to determine the resource which it points to
	nicTarget := source.Properties["attachedTo"]

	// don't draw dependency arrows to the attached resource
	if target.Id == nicTarget {
		return nil
	}

	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}
