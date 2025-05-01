package virtual_network

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
	TYPE   = types.VIRTUAL_NETWORK
	IMAGE  = images.VIRTUAL_NETWORK
	WIDTH  = 67
	HEIGHT = 40
)

var (
	STYLE = "fillColor=#dae8fc;strokeColor=#6c8ebf"
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

	vnetSize, ok := resource.Properties["size"]

	link := resource.GetLinkOrDefault()

	if !ok {
		return node.NewIcon(IMAGE, resource.Name, &geometry, link)
	}

	name := fmt.Sprintf("%s/%s", resource.Name, vnetSize[0])

	return node.NewIcon(IMAGE, name, &geometry, link)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{})
}

func (*handler) GroupResources(vnet *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInVnet := getAllResourcesInVnet(vnet.Id, resources, resource_map)

	if len(resourcesInVnet) == 0 {
		return nil
	}

	// a vnet can contain resources that belong to the same group, these needs to be filtered to
	// avoid moving the same group multiple times
	seenGroups := set.New[string]()

	resourcesInVnet = list.Filter(resourcesInVnet, func(n *node.Node) bool {
		if seenGroups.Contains(n.Id()) {
			return false
		}

		seenGroups.Add(n.Id())

		return true
	})

	vnetNode := (*resource_map)[vnet.Id].Node

	box := node.NewBox(&node.Geometry{
		X:      0,
		Y:      0,
		Width:  0,
		Height: 0,
	}, &STYLE)

	node.FillResourcesInBox(box, resourcesInVnet, diagram.Padding, true)

	vnetNode.SetProperty("parent", box.Id())
	vnetNode.ContainedIn = box
	node.SetIconRelativeTo(vnetNode, box, node.BOTTOM_LEFT)

	return []*node.Node{box}
}

func getAllResourcesInVnet(vnetId string, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	subnets := list.Filter(resources, func(r *models.Resource) bool {
		return list.Contains(r.DependsOn, func(dependency string) bool {
			return dependency == vnetId
		})
	})

	nodes := list.Map(subnets, func(r *models.Resource) *node.Node {
		return (*resource_map)[r.Id].Node.GetParentOrThis()
	})

	return nodes
}
