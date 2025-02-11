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
	// app service plans have an implicit dependency to a subnet. Don't draw these
	if target.Type == az.SUBNET {
		return nil
	}

	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) GroupResources(appServicePlan *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	resourcesInAppServicePlan := getResourcesInAppServicePlan(resources, appServicePlan.Id, resource_map)

	if len(resourcesInAppServicePlan) == 0 {
		return []*node.Node{}
	}

	firstAppServiceSubnet := getAppServiceSubnet(resourcesInAppServicePlan[0].Resource, resources)

	if firstAppServiceSubnet == nil {
		log.Printf("Could not determine subnet of app service plan %s", appServicePlan.Name)
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
	appServicePlanNode := (*resource_map)[appServicePlan.Id].Node
	appServicePlanNodeGeometry := appServicePlanNode.GetGeometry()

	box := node.NewBox(&node.Geometry{
		X:      appServicePlanNodeGeometry.X,
		Y:      appServicePlanNodeGeometry.Y,
		Width:  0,
		Height: 0,
	}, &STYLE)

	appServicePlanNode.SetProperty("parent", box.Id())
	appServicePlanNode.ContainedIn = box
	appServicePlanNode.SetPosition(0, 0)

	seenGroups := map[string]bool{}

	resourcesInAppServicePlan = list.Filter(resourcesInAppServicePlan, func(r *node.ResourceAndNode) bool {
		n := r.Node.GetParentOrThis()

		if seenGroups[n.Id()] {
			return false
		}

		seenGroups[n.Id()] = true

		return true
	})

	nodesToMove := list.Map(resourcesInAppServicePlan, func(r *node.ResourceAndNode) *node.Node {
		return r.Node.GetParentOrThis()
	})

	// move all resources in the app service plan into the box
	node.FillResourcesInBox(box, nodesToMove, diagram.Padding)

	node.ScaleDownAndSetIconBottomLeft(appServicePlanNode, box)

	// add an explicit dependency to the subnet
	appServicePlan.DependsOn = append(appServicePlan.DependsOn, *firstAppServiceSubnet)

	return []*node.Node{box}
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
