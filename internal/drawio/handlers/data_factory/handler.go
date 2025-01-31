package data_factory

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
)

type handler struct{}

const (
	TYPE   = az.DATA_FACTORY
	IMAGE  = images.DATA_FACTORY
	WIDTH  = 68
	HEIGHT = 68
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

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(dataFactory *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	resourcesInDataFactory := getResourcesInDataFactory(resources, dataFactory.Id, resource_map)

	dataFactoryNode := (*resource_map)[dataFactory.Id].Node
	dataFactoryNodeGeometry := dataFactoryNode.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      dataFactoryNodeGeometry.X,
		Y:      dataFactoryNodeGeometry.Y,
		Width:  200,
		Height: 200,
	}, nil)

	dataFactoryNode.SetProperty("parent", box.Id())
	dataFactoryNode.ContainedIn = box
	dataFactoryNode.SetPosition(0, 0)

	// move all resources in the adf into the box
	for _, resourceInAdf := range resourcesInDataFactory {
		resourceInAdf.Node.SetProperty("parent", box.Id())
		resourceInAdf.Node.ContainedIn = box
		resourceInAdf.Node.SetPosition(0, 0)
	}

	nodes = append(nodes, box)

	return nodes
}

func getResourcesInDataFactory(resources []*az.Resource, adfId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInAsp := list.Filter(resources, func(resource *az.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == adfId })
	})
	resourcesInAsp := list.Map(azResourcesInAsp, func(resource *az.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInAsp
}
