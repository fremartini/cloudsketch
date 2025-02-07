package private_endpoint

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.PRIVATE_ENDPOINT
	IMAGE  = images.PRIVATE_ENDPOINT
	WIDTH  = 68
	HEIGHT = 65
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

func (*handler) PostProcessIcon(privateEndpoint *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	// storage accounts might have multiple private endpoints attached to it
	attachedTo := (*resource_map)[privateEndpoint.Resource.Properties["attachedTo"]]

	if attachedTo.Resource.Type == az.WEB_SITES {
		addImplicitDependencyToFunctionApp(privateEndpoint.Resource, attachedTo.Resource, resource_map)
	}

	attachedPrivateEndpoints := getPrivateEndpointPointingToResource(resource_map, attachedTo.Resource)

	// multiple private endpoints point to this resource
	if len(attachedPrivateEndpoints) > 1 {
		return nil
	}

	// set icon top right
	return node.SetIcon(attachedTo.Node, privateEndpoint.Node, node.TOP_RIGHT)
}

func addImplicitDependencyToFunctionApp(privateEndpoint, functionApp *az.Resource, resource_map *map[string]*node.ResourceAndNode) {
	// App Service Plans need a reference to the subnet it should be added to. This is fetched from the
	// resources inside the plan. If the resource this Private Endpoint is attached to, is a function app
	// an implicit dependency is added to the App Service to reference
	for _, dependency := range privateEndpoint.DependsOn {
		dependentResource := (*resource_map)[dependency]

		if dependentResource == nil {
			continue
		}

		if dependentResource.Resource.Type != az.SUBNET {
			continue
		}

		// dependency is a subnet. Add it to the function app
		functionApp.DependsOn = append(functionApp.DependsOn, dependency)
	}
}

func getPrivateEndpointPointingToResource(resource_map *map[string]*node.ResourceAndNode, attachedResource *az.Resource) []*az.Resource {
	privateEndpoints := []*az.Resource{}

	// figure out how many private endpoints are pointing to the storage account
	for _, v := range *resource_map {
		// filter out the private endpoints
		if v.Resource.Type != az.PRIVATE_ENDPOINT {
			continue
		}

		if v.Resource.Properties["attachedTo"] != attachedResource.Id {
			continue
		}

		// another private endpoints point to the same resource
		if (*resource_map)[v.Resource.Id].Node != nil {
			privateEndpoints = append(privateEndpoints, v.Resource)
		}
	}

	return privateEndpoints
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	// don't draw arrows to subnets
	if target.Type == az.SUBNET {
		return nil
	}

	sourceNode := (*resource_map)[source.Id].Node
	targetNode := (*resource_map)[target.Id].Node

	// if they are in the same group, don't draw the arrow
	if sourceNode.ContainedIn != nil && targetNode.ContainedIn != nil {
		if sourceNode.ContainedIn == targetNode.ContainedIn {
			return nil
		}
	}

	return node.NewArrow(sourceNode.Id(), targetNode.Id())
}

func (*handler) DrawBox(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
