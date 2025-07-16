package private_dns_zone

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
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
	// don't draw arrows to subscriptions
	typeBlacklist := []string{types.SUBSCRIPTION, types.MANAGEMENT_GROUP}

	// remove entries from the blacklist
	targets = list.Filter(targets, func(target *models.Resource) bool {
		return !list.Contains(typeBlacklist, func(t string) bool {
			return target.Type == t
		})
	})

	targetResources := list.Map(targets, func(target *models.Resource) *node.ResourceAndNode {
		return (*resource_map)[target.Id]
	})

	sourceNode := (*resource_map)[source.Id].Node

	arrows := list.Fold(targetResources, []*node.Arrow{}, func(target *node.ResourceAndNode, acc []*node.Arrow) []*node.Arrow {
		return append(acc, node.NewArrow(sourceNode.Id(), target.Node.Id(), nil))
	})

	return arrows
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
