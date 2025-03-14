package private_dns_zone

import (
	"cloudsketch/internal/drawio/handlers/diagram"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
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

func (*handler) DrawDependency(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	arrows := []*node.Arrow{}

	sourceNode := (*resource_map)[source.Id].Node

	for _, target := range targets {

		targetNode := (*resource_map)[target.Id].Node

		// if they are in the same group, don't draw the arrow
		if sourceNode.ContainedIn != nil && targetNode.ContainedIn != nil {
			if sourceNode.GetParentOrThis() == targetNode.GetParentOrThis() {
				continue
			}
		}

		arrows = append(arrows, node.NewArrow(sourceNode.Id(), targetNode.Id(), nil))
	}

	return arrows
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
	node.FillResourcesInBox(box, nodesToMove, diagram.Padding)

	node.ScaleDownAndSetIconRelativeTo(privateDNSZoneNode, box, node.BOTTOM_LEFT)

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
