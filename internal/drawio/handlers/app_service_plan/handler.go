package app_service_plan

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/diagram"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
	"log"
)

type handler struct{}

const (
	TYPE   = az.APP_SERVICE_PLAN
	IMAGE  = images.APP_SERVICE_PLAN
	WIDTH  = 64
	HEIGHT = 64
)

var (
	STYLE = "rounded=0;whiteSpace=wrap;html=1;dashed=1;opacity=50;"
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource) *node.Node {
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
	// app service plans have an implicit dependency to a subnet. Don't draw these
	if target.Type == az.SUBNET {
		return nil
	}

	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(appService *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInAppServicePlan := getResourcesInAppServicePlan(resources, appService.Id, resource_map)

	if len(resourcesInAppServicePlan) == 0 {
		return []*node.Node{}
	}

	firstAppServiceSubnet := getAppServiceSubnet(resourcesInAppServicePlan[0].Resource, resources)

	if firstAppServiceSubnet == nil {
		log.Printf("Could not determine subnet of app service plan %s", appService.Name)
		return []*node.Node{}
	}

	// if all app services in the plan belong to the same subnet a box can be draw
	allAppServicesInSameSubnet := list.Fold(resourcesInAppServicePlan[1:], true, func(resource *node.ResourceAndNode, matches bool) bool {
		appServiceSubnet := getAppServiceSubnet(resource.Resource, resources)

		return matches && appServiceSubnet == firstAppServiceSubnet
	})

	if !allAppServicesInSameSubnet {
		return []*node.Node{}
	}

	// draw the box
	appServiceNode := (*resource_map)[appService.Id].Node
	appServiceNodeGeometry := appServiceNode.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      appServiceNodeGeometry.X,
		Y:      appServiceNodeGeometry.Y,
		Width:  0,
		Height: 0,
	}, &STYLE)

	appServiceNode.SetProperty("parent", box.Id())
	appServiceNode.ContainedIn = box
	appServiceNode.SetPosition(0, 0)

	// move all resources in the app service plan into the box
	node.FillResourcesInBox(box, resourcesInAppServicePlan, diagram.Padding)

	node.ScaleDownAndSetIconBottomLeft(appServiceNode, box)

	// add an implicit dependency to the subnet
	appService.DependsOn = append(appService.DependsOn, *firstAppServiceSubnet)

	nodes := []*node.Node{}

	nodes = append(nodes, box)

	return nodes
}

func getResourcesInAppServicePlan(resources []*az.Resource, aspId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInAsp := list.Filter(resources, func(resource *az.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency string) bool { return dependency == aspId })
	})
	resourcesInAsp := list.Map(azResourcesInAsp, func(resource *az.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInAsp
}

func getAppServiceSubnet(appService *az.Resource, resources []*az.Resource) *string {
	for _, dependency := range appService.DependsOn {
		resource := list.First(resources, func(resource *az.Resource) bool {
			return resource.Id == dependency
		})

		if resource.Type == az.SUBNET {
			return &resource.Id
		}
	}

	return nil
}
