package app_service_plan

import (
	"cloudsketch/internal/datastructures/set"
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
	"log"
)

type handler struct{}

const (
	TYPE   = types.APP_SERVICE_PLAN
	IMAGE  = images.APP_SERVICE_PLAN
	WIDTH  = 64
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
	// app service plans have an implicit dependency to a subnet. Don't draw these
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{types.SUBNET})
}

func (*handler) GroupResources(appServicePlan *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
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

		return matches && appServiceSubnet.Id == firstAppServiceSubnet.Id
	})

	if !allAppServicesInSameSubnet {
		return []*node.Node{}
	}

	seenGroups := set.New[string]()

	resourcesInAppServicePlan = list.Filter(resourcesInAppServicePlan, func(r *node.ResourceAndNode) bool {
		n := r.Node.GetParentOrThis()

		if seenGroups.Contains(n.Id()) {
			return false
		}

		seenGroups.Add(n.Id())

		return true
	})

	appServicePlanNode := (*resource_map)[appServicePlan.Id].Node

	box := node.BoxResources(appServicePlanNode, resourcesInAppServicePlan)

	// add an explicit dependency to the subnet
	appServicePlan.DependsOn = append(appServicePlan.DependsOn, firstAppServiceSubnet)

	return []*node.Node{box}
}

func getResourcesInAppServicePlan(resources []*models.Resource, aspId string, resource_map *map[string]*node.ResourceAndNode) []*node.ResourceAndNode {
	azResourcesInAsp := list.Filter(resources, func(resource *models.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency *models.Resource) bool { return dependency.Id == aspId })
	})
	resourcesInAsp := list.Map(azResourcesInAsp, func(resource *models.Resource) *node.ResourceAndNode {
		return (*resource_map)[resource.Id]
	})
	return resourcesInAsp
}

func getAppServiceSubnet(appService *models.Resource, resources []*models.Resource) *models.Resource {
	for _, dependency := range appService.DependsOn {
		resource := list.First(resources, func(resource *models.Resource) bool {
			return resource.Id == dependency.Id
		})

		if resource.Type == types.SUBNET {
			return resource
		}
	}

	return nil
}
