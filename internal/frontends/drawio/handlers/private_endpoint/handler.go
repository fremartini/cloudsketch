package private_endpoint

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.PRIVATE_ENDPOINT
	IMAGE  = images.PRIVATE_ENDPOINT
	WIDTH  = 68
	HEIGHT = 65
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

func (*handler) PostProcessIcon(privateEndpoint *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	attachedToIds, ok := privateEndpoint.Resource.Properties["attachedTo"]

	if !ok {
		return nil
	}

	// storage accounts might have multiple private endpoints attached to it
	attachedTo, ok := (*resource_map)[attachedToIds[0]]

	// the attached resource was not rendered likely because its in another subscription
	if !ok {
		return nil
	}

	if attachedTo.Resource.Type == types.APP_SERVICE ||
		attachedTo.Resource.Type == types.FUNCTION_APP ||
		attachedTo.Resource.Type == types.LOGIC_APP {
		addImplicitDependencyToFunctionApp(privateEndpoint.Resource, attachedTo.Resource, resource_map)
	}

	attachedPrivateEndpoints := getPrivateEndpointPointingToResource(resource_map, attachedTo.Resource)

	if len(attachedPrivateEndpoints) > 1 {
		// multiple private endpoints point to this resource. If they all
		// belong to the same subnet they can be merged
		resources := []*models.Resource{}
		for _, e := range *resource_map {
			resources = append(resources, e.Resource)
		}

		firstSubnet := getPrivateEndpointSubnet(privateEndpoint.Resource, resources)

		allPrivateEndpointsInSameSubnet := list.Fold(attachedPrivateEndpoints, true, func(resource *models.Resource, matches bool) bool {
			privateEndpointSubnet := getPrivateEndpointSubnet(resource, resources)

			return matches && privateEndpointSubnet == firstSubnet
		})

		if !allPrivateEndpointsInSameSubnet {
			return nil
		}

		// delete unneeded private endpoint icons
		for _, pe := range attachedPrivateEndpoints {
			if pe.Id == privateEndpoint.Resource.Id {
				continue
			}

			delete(*resource_map, pe.Id)
		}
	}

	// one private endpoint exists, "merge" the two icons
	return node.GroupIconsAndSetPosition(attachedTo.Node, privateEndpoint.Node, node.TOP_RIGHT)
}

func getPrivateEndpointSubnet(resource *models.Resource, resources []*models.Resource) *string {
	for _, dependency := range resource.DependsOn {
		resource := list.FirstOrDefault(resources, nil, func(resource *models.Resource) bool {
			return resource.Id == dependency.Id
		})

		if resource == nil {
			continue
		}

		if resource.Type == types.SUBNET {
			return &resource.Id
		}
	}

	return nil
}

func addImplicitDependencyToFunctionApp(privateEndpoint, functionApp *models.Resource, resource_map *map[string]*node.ResourceAndNode) {
	// App Service Plans need a reference to the subnet it should be added to. This is fetched from the
	// resources inside the plan. If the resource this Private Endpoint is attached to, is a function app
	// an implicit dependency is added to the App Service to reference
	for _, dependency := range privateEndpoint.DependsOn {
		dependentResource := (*resource_map)[dependency.Id]

		if dependentResource == nil {
			continue
		}

		if dependentResource.Resource.Type != types.SUBNET {
			continue
		}

		// dependency is a subnet. Add it to the function app
		functionApp.DependsOn = append(functionApp.DependsOn, dependency)
	}
}

func getPrivateEndpointPointingToResource(resource_map *map[string]*node.ResourceAndNode, attachedResource *models.Resource) []*models.Resource {
	privateEndpoints := []*models.Resource{}

	// figure out how many private endpoints are pointing to the storage account
	for _, v := range *resource_map {
		// filter out the private endpoints
		if v.Resource.Type != types.PRIVATE_ENDPOINT {
			continue
		}

		attachedToIds, ok := v.Resource.Properties["attachedTo"]

		if !ok {
			continue
		}

		if attachedToIds[0] != attachedResource.Id {
			continue
		}

		// another private endpoints point to the same resource
		if (*resource_map)[v.Resource.Id].Node != nil {
			privateEndpoints = append(privateEndpoints, v.Resource)
		}
	}

	return privateEndpoints
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{types.SUBNET})
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
