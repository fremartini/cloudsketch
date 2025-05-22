package api_management_service

import (
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/frontends/drawio/handlers/diagram"
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.API_MANAGEMENT_SERVICE
	IMAGE  = images.API_MANAGEMENT_SERVICE
	WIDTH  = 65
	HEIGHT = 60
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

func (*handler) GroupResources(resource *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInAPIM := node.GetChildResourcesOfType(resources, resource.Id, types.API_MANAGEMENT_API, resource_map)

	if len(resourcesInAPIM) == 0 {
		return []*node.Node{}
	}

	// draw the box
	apimNode := (*resource_map)[resource.Id].Node
	apimNodeGeometry := apimNode.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      apimNodeGeometry.X,
		Y:      apimNodeGeometry.Y,
		Width:  0,
		Height: 0,
	}, &STYLE)

	apimNode.SetProperty("parent", box.Id())
	apimNode.ContainedIn = box
	apimNode.SetPosition(0, 0)

	seenGroups := set.New[string]()

	resourcesInAPIM = list.Filter(resourcesInAPIM, func(r *node.ResourceAndNode) bool {
		n := r.Node.GetParentOrThis()

		if seenGroups.Contains(n.Id()) {
			return false
		}

		seenGroups.Add(n.Id())

		return true
	})

	nodesToMove := list.Map(resourcesInAPIM, func(r *node.ResourceAndNode) *node.Node {
		return r.Node.GetParentOrThis()
	})

	// move all resources in the app service plan into the box
	node.FillResourcesInBox(box, nodesToMove, diagram.Padding, true)

	apimNode.SetDimensions(apimNodeGeometry.Width/2, apimNodeGeometry.Height/2)
	node.SetIconRelativeTo(apimNode, box, node.BOTTOM_LEFT)

	return []*node.Node{box}
}
