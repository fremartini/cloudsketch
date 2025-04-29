package private_dns_zone

import (
	"cloudsketch/internal/frontends/drawio/handlers/diagram"
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/drawio/models"
	"cloudsketch/internal/frontends/drawio/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.PRIVATE_DNS_ZONE
	IMAGE  = images.PRIVATE_DNS_ZONE
	WIDTH  = 64
	HEIGHT = 64
)

var (
	STYLE = "rounded=0;whiteSpace=wrap;html=1;dashed=1;opacity=50;"
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
	resourcesInPrivateDNSZone := getResourcesInPrivateDNSZone(resources, privateDNSZone.Id, resource_map)

	if len(resourcesInPrivateDNSZone) == 0 {
		return []*node.Node{}
	}

	privateDNSZoneNode := (*resource_map)[privateDNSZone.Id].Node
	privateDNSZoneNodeGeometry := privateDNSZoneNode.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      privateDNSZoneNodeGeometry.X,
		Y:      privateDNSZoneNodeGeometry.Y,
		Width:  0,
		Height: 0,
	}, &STYLE)

	privateDNSZoneNode.SetProperty("parent", box.Id())
	privateDNSZoneNode.ContainedIn = box
	privateDNSZoneNode.SetPosition(0, 0)

	nodesToMove := list.Map(resourcesInPrivateDNSZone, func(r *node.ResourceAndNode) *node.Node {
		return r.Node.GetParentOrThis()
	})

	// move all resources in the private dns zone into the box
	node.FillResourcesInBox(box, nodesToMove, diagram.Padding, true)

	privateDNSZoneNode.SetDimensions(privateDNSZoneNodeGeometry.Width/2, privateDNSZoneNodeGeometry.Height/2)
	node.SetIconRelativeTo(privateDNSZoneNode, box, node.BOTTOM_LEFT)

	return []*node.Node{box}
}

func getResourcesInPrivateDNSZone(resources []*models.Resource, adfId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInAsp := list.Filter(resources, func(resource *models.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == adfId })
	})
	resourcesInAsp := list.Map(azResourcesInAsp, func(resource *models.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInAsp
}
