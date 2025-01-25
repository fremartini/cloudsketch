package app_service_plan

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
	"sort"
)

type handler struct{}

const (
	TYPE   = az.APP_SERVICE_PLAN
	IMAGE  = images.APP_SERVICE_PLAN
	WIDTH  = 64
	HEIGHT = 64
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

func (*handler) DrawBox(resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	nodes := []*node.Node{}

	appServicesToProcess := list.Filter(resources, func(resource *az.Resource) bool { return resource.Type == az.APP_SERVICE_PLAN })

	// ensure some deterministic order
	sort.Slice(appServicesToProcess, func(i, j int) bool {
		return appServicesToProcess[i].Name < appServicesToProcess[j].Name
	})

	for _, appService := range appServicesToProcess {
		resourcesInAppServicePlan := getResourcesInAppServicePlan(resources, appService.Id, resource_map)

		if len(resourcesInAppServicePlan) == 0 {
			continue
		}

		firstAppServiceSubnet := getAppServiceSubnet(resourcesInAppServicePlan[0].Resource, resources)

		// if all app services in the plan belong to the same subnet a box can be draw
		allAppServicesInSameSubnet := list.Fold(resourcesInAppServicePlan[1:], true, func(resource *node.ResourceAndNode, matches bool) bool {
			appServiceSubnet := getAppServiceSubnet(resource.Resource, resources)

			return matches && appServiceSubnet == firstAppServiceSubnet
		})

		if !allAppServicesInSameSubnet {
			continue
		}

		// TODO: draw the box
	}

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
