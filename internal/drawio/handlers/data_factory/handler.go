package data_factory

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/diagram"
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

func (*handler) GroupResources(dataFactory *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	resourcesInDataFactory := getResourcesInDataFactory(resources, dataFactory.Id, resource_map)

	dataFactoryNode := (*resource_map)[dataFactory.Id].Node
	dataFactoryNodeGeometry := dataFactoryNode.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      dataFactoryNodeGeometry.X,
		Y:      dataFactoryNodeGeometry.Y,
		Width:  0,
		Height: 0,
	}, &STYLE)

	dataFactoryNode.SetProperty("parent", box.Id())
	dataFactoryNode.ContainedIn = box
	dataFactoryNode.SetPosition(0, 0)

	nodesToMove := list.Map(resourcesInDataFactory, func(r *node.ResourceAndNode) *node.Node {
		return r.Node.GetParentOrThis()
	})

	// move all resources in the adf into the box
	node.FillResourcesInBox(box, nodesToMove, diagram.Padding)

	node.ScaleDownAndSetIconBottomLeft(dataFactoryNode, box)

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
