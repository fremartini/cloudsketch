package private_dns_zone

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/diagram"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
)

type handler struct{}

const (
	TYPE   = az.PRIVATE_DNS_ZONE
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

func (*handler) MapResource(resource *az.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	return node.NewIcon(IMAGE, resource.Name, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) GroupResources(privateDNSZone *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	resourcesInPrivateDNSZone := getResourcesInPrivateDNSZone(resources, privateDNSZone.Id, resource_map)

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
	node.FillResourcesInBox(box, nodesToMove, diagram.Padding)

	node.ScaleDownAndSetIconBottomLeft(privateDNSZoneNode, box)

	nodes = append(nodes, box)

	return nodes
}

func getResourcesInPrivateDNSZone(resources []*az.Resource, adfId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInAsp := list.Filter(resources, func(resource *az.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == adfId })
	})
	resourcesInAsp := list.Map(azResourcesInAsp, func(resource *az.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInAsp
}
