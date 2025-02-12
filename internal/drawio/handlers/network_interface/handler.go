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

func (*handler) MapResource(resource *az.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	return node.NewIcon(IMAGE, resource.Name, &geometry)
}

func (*handler) PostProcessIcon(nic *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	attachedTo := (*resource_map)[nic.Resource.Properties["attachedTo"]]

	// dont draw NICs if they are attached to a blacklisted resource
	if attachedTo == nil || isBlacklistedResource(attachedTo.Resource.Type) {
		delete(*resource_map, nic.Resource.Id)
		return nil
	}

	existingNics := getNICsPointingToResource(resource_map, attachedTo.Resource)

	// multiple NICs point to the same resource - skip
	if len(existingNics) > 1 {
		return nil
	}

	// set icon top right
	return node.SetIcon(attachedTo.Node, nic.Node, node.TOP_RIGHT)
}

func isBlacklistedResource(resourceType string) bool {
	blacklist := []string{az.PRIVATE_ENDPOINT}

	return list.Contains(blacklist, func(e string) bool {
		return resourceType == e
	})
}

func getNICsPointingToResource(resource_map *map[string]*node.ResourceAndNode, attachedResource *az.Resource) []*az.Resource {
	nics := []*az.Resource{}

	// figure out how many private endpoints are pointing to the storage account
	for _, v := range *resource_map {
		// filter out the private endpoints
		if v.Resource.Type != az.NETWORK_INTERFACE {
			continue
		}

		if v.Resource.Properties["attachedTo"] != attachedResource.Id {
			continue
		}

		// another private endpoints point to the same resource
		if (*resource_map)[v.Resource.Id].Node != nil {
			nics = append(nics, v.Resource)
		}
	}

	return nics
}

func (*handler) DrawDependency(source *az.Resource, targets []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	arrows := []*node.Arrow{}

	sourceNode := (*resource_map)[source.Id].Node

	for _, target := range targets {
		// don't draw arrows to subnets
		if target.Type == az.SUBNET {
			continue
		}

		targetNode := (*resource_map)[target.Id].Node

		arrows = append(arrows, node.NewArrow(sourceNode.Id(), targetNode.Id(), nil))
	}

	return arrows
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
