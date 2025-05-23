package private_dns_zone

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
)

type handler struct{}

const (
	TYPE   = types.PRIVATE_DNS_ZONE
	IMAGE  = images.PRIVATE_DNS_ZONE
	WIDTH  = 64
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
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{})
}

func (*handler) GroupResources(privateDNSZone *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInPrivateDNSZone := node.GetChildResourcesOfType(resources, privateDNSZone.Id, types.DNS_RECORD, resource_map)

	if len(resourcesInPrivateDNSZone) == 0 {
		return []*node.Node{}
	}

	privateDNSZoneNode := (*resource_map)[privateDNSZone.Id].Node

	box := node.BoxResources(privateDNSZoneNode, resourcesInPrivateDNSZone)

	return []*node.Node{box}
}
