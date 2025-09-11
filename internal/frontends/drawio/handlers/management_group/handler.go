package management_group

import (
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/frontends/drawio/handlers/diagram"
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
	"fmt"
)

type handler struct{}

const (
	TYPE   = types.MANAGEMENT_GROUP
	IMAGE  = images.MANAGEMENT_GROUP
	WIDTH  = 66
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
	return node.DrawDependencyArrowsToTargets(source, targets, resource_map, []string{})
}

func (*handler) GroupResources(resource *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	fmt.Println(resource.Name)

	managementGroupResources := getAllResourcesInManagementGroup(resource.Id, resources, resource_map)

	if len(managementGroupResources) == 0 {
		return []*node.Node{}
	}

	// a management group can contain resources that belong to the same group, these needs to be filtered to
	// avoid moving the same group multiple times
	seenGroups := set.New[string]()

	managementGroupResources = list.Filter(managementGroupResources, func(n *node.Node) bool {
		if seenGroups.Contains(n.Id()) {
			return false
		}

		seenGroups.Add(n.GetParentOrThis().Id())

		return true
	})

	managementGroupNode := (*resource_map)[resource.Id].Node

	box := node.NewBox(&node.Geometry{
		X:      0,
		Y:      0,
		Width:  0,
		Height: 0,
	}, nil)

	node.FillResourcesInBox(box, managementGroupResources, diagram.Padding, true)

	managementGroupNode.SetProperty("parent", box.Id())
	managementGroupNode.ContainedIn = box
	node.SetIconRelativeTo(managementGroupNode, box, node.TOP_LEFT)

	return []*node.Node{box}
}

func getAllResourcesInManagementGroup(resourceId string, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	managementGroupResources := list.Filter(resources, func(r *models.Resource) bool {
		return list.Contains(r.DependsOn, func(dependency *models.Resource) bool {
			return dependency.Id == resourceId
		})
	})

	nodes := list.Map(managementGroupResources, func(r *models.Resource) *node.Node {
		return (*resource_map)[r.Id].Node.GetParentOrThis()
	})

	return nodes
}
