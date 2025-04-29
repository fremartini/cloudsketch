package network_interface

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/drawio/models"
	"cloudsketch/internal/frontends/drawio/types"
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
	attachedTo, ok := (*resource_map)[nic.Resource.Properties["attachedTo"][0]]

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

	// set icon top right
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

	// figure out how many private endpoints are pointing to the storage account
	for _, v := range *resource_map {
		// filter out the private endpoints
		if v.Resource.Type != types.NETWORK_INTERFACE {
			continue
		}

		if v.Resource.Properties["attachedTo"][0] != attachedResource.Id {
			continue
		}

		// another private endpoints point to the same resource
		if (*resource_map)[v.Resource.Id].Node != nil {
			nics = append(nics, v.Resource)
		}
	}

	return nics
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{types.SUBNET})
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
