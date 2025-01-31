package private_dns_zone

import (
	"azsample/internal/az"
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

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	n := node.NewIcon(IMAGE, resource.Name, &geometry)

	return []*node.Node{n}
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(privateDNSZone *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	resourcesInDataFactory := getResourcesInPrivateDNSZone(resources, privateDNSZone.Id, resource_map)

	privateDNSZoneNode := (*resource_map)[privateDNSZone.Id].Node
	privateDNSZoneNodeGeometry := privateDNSZoneNode.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      privateDNSZoneNodeGeometry.X,
		Y:      privateDNSZoneNodeGeometry.Y,
		Width:  200,
		Height: 200,
	}, nil)

	privateDNSZoneNode.SetProperty("parent", box.Id())
	privateDNSZoneNode.ContainedIn = box
	privateDNSZoneNode.SetPosition(0, 0)

	// move all resources in the private dns zone into the box
	for _, resourceInAdf := range resourcesInDataFactory {
		resourceInAdf.Node.SetProperty("parent", box.Id())
		resourceInAdf.Node.ContainedIn = box
		resourceInAdf.Node.SetPosition(0, 0)
	}

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
