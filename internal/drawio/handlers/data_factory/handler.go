package data_factory

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
	TYPE   = types.DATA_FACTORY
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

func (*handler) GroupResources(dataFactory *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInDataFactory := getResourcesInDataFactory(resources, dataFactory.Id, resource_map)

	// the resouces in the data factory can include its private endpoint, this needs to be handled differently
	resourcesInDataFactory, attachedResources := list.Split(resourcesInDataFactory, func(ran *node.ResourceAndNode) bool {
		if ran.Resource.Type != types.PRIVATE_ENDPOINT {
			return true
		}

		attachedTo, ok := ran.Resource.Properties["attachedTo"]

		if !ok || attachedTo != dataFactory.Id {
			return true
		}

		return attachedTo != dataFactory.Id
	})

	dataFactoryNode := (*resource_map)[dataFactory.Id].Node
	dataFactoryGroup := dataFactoryNode.GetParentOrThis()
	dataFactoryGroupGeometry := dataFactoryGroup.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      dataFactoryGroupGeometry.X,
		Y:      dataFactoryGroupGeometry.Y,
		Width:  0,
		Height: 0,
	}, &STYLE)

	if len(attachedResources) > 0 {
		dataFactoryNode.SetDimensions(dataFactoryNode.GetGeometry().Width/2, dataFactoryNode.GetGeometry().Height/2)

		for _, attachedResource := range attachedResources {
			attachedResource.Node.SetDimensions(attachedResource.Node.GetGeometry().Width/2, attachedResource.Node.GetGeometry().Height/2)
			node.SetIconRelativeTo(attachedResource.Node, dataFactoryNode, node.TOP_RIGHT)
		}
	}

	dataFactoryGroup.SetDimensions(dataFactoryGroupGeometry.Width/2, dataFactoryGroupGeometry.Height/2)

	dataFactoryGroup.SetProperty("parent", box.Id())
	dataFactoryGroup.ContainedIn = box
	dataFactoryGroup.SetPosition(0, 0)

	nodesToMove := list.Map(resourcesInDataFactory, func(r *node.ResourceAndNode) *node.Node {
		return r.Node.GetParentOrThis()
	})

	// move all resources in the adf into the box
	node.FillResourcesInBox(box, nodesToMove, diagram.Padding, true)
	node.SetIconRelativeTo(dataFactoryGroup, box, node.BOTTOM_LEFT)

	return []*node.Node{box}
}

func getResourcesInDataFactory(resources []*models.Resource, adfId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInAsp := list.Filter(resources, func(resource *models.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == adfId })
	})
	resourcesInAsp := list.Map(azResourcesInAsp, func(resource *models.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInAsp
}
