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

func addImplicitDependencyToFunctionApp(private_endpoint, attachedResource *az.Resource, resource_map *map[string]*node.ResourceAndNode) {
	// App Service Plans need a reference to the subnet it should be added to. This is fetched from the
	// resources inside the plan. If the resource this Private Endpoint is attached to, is a function app
	// an implicit dependency is added to the App Service Plan to reference
	for _, dependency := range private_endpoint.DependsOn {
		dependentResource := (*resource_map)[dependency]

		if dependentResource == nil {
			continue
		}

		if dependentResource.Resource.Type != az.SUBNET {
			continue
		}

		if attachedResource.Type != az.WEB_SITES {
			continue
		}

		attachedResource.DependsOn = append(attachedResource.DependsOn, dependency)
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

func (*handler) PostProcessIcon(private_endpoint *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	// storage accounts might have multiple private endpoints attached to it
	attachedTo := (*resource_map)[private_endpoint.Resource.Properties["attachedTo"]]

	addImplicitDependencyToFunctionApp(private_endpoint.Resource, attachedTo.Resource, resource_map)

	existingPrivateEndpoints := getPrivateEndpointPointingToResource(resource_map, attachedTo.Resource)

	// multiple private endpoints point to the same resource - skip
	if len(existingPrivateEndpoints) > 1 {
		return nil
	}

	// set icon top right
	return node.SetIcon(attachedTo.Node, private_endpoint.Node, resource_map, node.TOP_RIGHT)
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
