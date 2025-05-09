package subscription

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
	TYPE   = types.SUBSCRIPTION
	IMAGE  = images.SUBSCRIPTION
	WIDTH  = 68
	HEIGHT = 68
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
	subscriptionResources := getAllResourcesInSubscription(resource.Id, resources, resource_map)

	// a subscription can contain resources that belong to the same group, these needs to be filtered to
	// avoid moving the same group multiple times
	seenGroups := set.New[string]()

	subscriptionResources = list.Filter(subscriptionResources, func(n *node.Node) bool {
		if seenGroups.Contains(n.Id()) {
			return false
		}

		seenGroups.Add(n.Id())

		return true
	})

	subscriptionNode := (*resource_map)[resource.Id].Node

	box := node.NewBox(&node.Geometry{
		X:      0,
		Y:      0,
		Width:  0,
		Height: 0,
	}, nil)

	node.FillResourcesInBox(box, subscriptionResources, diagram.Padding, false)

	subscriptionNode.SetProperty("parent", box.Id())
	subscriptionNode.ContainedIn = box
	node.SetIconRelativeTo(subscriptionNode, box, node.TOP_LEFT)

	return []*node.Node{box}
}

func getAllResourcesInSubscription(resourceId string, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	subscriptionResources := list.Filter(resources, func(r *models.Resource) bool {
		return list.Contains(r.DependsOn, func(dependency *models.Resource) bool {
			return dependency.Id == resourceId
		})
	})

	nodes := list.Map(subscriptionResources, func(r *models.Resource) *node.Node {
		return (*resource_map)[r.Id].Node.GetParentOrThis()
	})

	return nodes
}
