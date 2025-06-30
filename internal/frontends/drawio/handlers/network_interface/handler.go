package network_interface

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.NETWORK_INTERFACE
	IMAGE  = images.NETWORK_INTERFACE
	WIDTH  = 68
	HEIGHT = 60
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

func (*handler) PostProcessIcon(nic *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	attachedToIds, ok := nic.Resource.Properties["attachedTo"]

	if !ok {
		return nil
	}

	attachedTo, ok := (*resource_map)[attachedToIds[0]]

	// dont draw NICs if they are attached to a blacklisted resource
	if !ok || isBlacklistedResource(attachedTo.Resource.Type) {
		delete(*resource_map, nic.Resource.Id)
		return nil
	}

	existingNics := getNICsPointingToResource(resource_map, attachedTo.Resource)

	// multiple NICs point to the same resource - skip
	if len(existingNics) > 1 {
		return nil
	}

	return node.GroupIconsAndSetPosition(attachedTo.Node, nic.Node, node.TOP_RIGHT)
}

func isBlacklistedResource(resourceType string) bool {
	blacklist := []string{types.PRIVATE_ENDPOINT}

	return list.Contains(blacklist, func(e string) bool {
		return resourceType == e
	})
}

func getNICsPointingToResource(resource_map *map[string]*node.ResourceAndNode, attachedResource *models.Resource) []*models.Resource {
	nics := []*models.Resource{}

	// figure out how many private endpoints are pointing to this resource
	for _, v := range *resource_map {
		if v.Resource.Type != types.NETWORK_INTERFACE {
			continue
		}

		attachedTo, ok := v.Resource.Properties["attachedTo"]

		if !ok {
			continue
		}

		if attachedTo[0] != attachedResource.Id {
			continue
		}

		// another private endpoints point to the same resource
		if (*resource_map)[v.Resource.Id].Node != nil {
			nics = append(nics, v.Resource)
		}
	}

	return nics
}

func (*handler) DrawDependencies(nic *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	arrows := node.DrawDependencyArrowsToTargets(nic, targets, resource_map, []string{})

	attachedToIds, ok := nic.Properties["attachedTo"]

	if !ok {
		return arrows
	}

	attachedTo, ok := (*resource_map)[attachedToIds[0]]

	if !ok {
		return arrows
	}

	existingNics := getNICsPointingToResource(resource_map, attachedTo.Resource)

	// multiple NICs point to the same resource
	if len(existingNics) > 1 {
		thisResource := (*resource_map)[nic.Id].Node

		arrowToAttachedResource := node.NewArrow(thisResource.Id(), attachedTo.Node.Id(), nil)

		arrows = append(arrows, arrowToAttachedResource)
	}

	return arrows
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
