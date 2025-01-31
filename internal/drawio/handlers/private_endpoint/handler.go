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

func (*handler) DrawIcon(private_endpoint *az.Resource, resources *map[string]*node.ResourceAndNode) []*node.Node {
	// private endpoint is not attached
	attachedResourceId, ok := private_endpoint.Properties["attachedTo"]

	if !ok {
		return []*node.Node{}
	}

	attachedResource := (*resources)[attachedResourceId]

	// the linked resource has not been drawn
	if attachedResource == nil || attachedResource.Node == nil {
		return []*node.Node{}
	}

	// storage accounts might have multiple private endpoints attached to it
	if shouldExit := isOtherPrivateEndpointPointingToTheSameResource(private_endpoint, resources, attachedResource.Resource); shouldExit {
		geometry := node.Geometry{
			X:      0,
			Y:      0,
			Width:  WIDTH,
			Height: HEIGHT,
		}

		n := node.NewIcon(IMAGE, private_endpoint.Name, &geometry)

		return []*node.Node{n}
	}

	addImplicitDependencyToFunctionApp(private_endpoint, resources, attachedResource.Resource)

	return node.SetIcon(attachedResource, resources, IMAGE, HEIGHT, WIDTH, node.TOP_RIGHT)
}

func addImplicitDependencyToFunctionApp(private_endpoint *az.Resource, resources *map[string]*node.ResourceAndNode, attachedResource *az.Resource) {
	// App Service Plans need a reference to the subnet it should be added to. This is fetched from the
	// resources inside the plan. If the resource this Private Endpoint is attached to, is a function app
	// an implicit dependency is added to the App Service Plan to reference
	for _, dependency := range private_endpoint.DependsOn {
		dependentResource := (*resources)[dependency].Resource

		if dependentResource.Type != az.SUBNET {
			continue
		}

		if attachedResource.Type != az.FUNCTION_APP {
			continue
		}

		attachedResource.DependsOn = append(attachedResource.DependsOn, dependency)
	}
}

func isOtherPrivateEndpointPointingToTheSameResource(resource *az.Resource, resources *map[string]*node.ResourceAndNode, linkedResource *az.Resource) bool {
	if linkedResource.Type != az.STORAGE_ACCOUNT {
		return false
	}

	// figure out if there are other private endpoints pointed to the storage account
	for _, v := range *resources {
		// filter out the private endpoints
		if v.Resource.Type != az.PRIVATE_ENDPOINT {
			continue
		}

		// filter out this resource
		if v.Resource.Id == resource.Id {
			continue
		}

		if v.Resource.Properties["attachedTo"] != linkedResource.Id {
			continue
		}

		// another private endpoints point to the same resource. If it has been rendered, don't render this one
		if (*resources)[v.Resource.Id].Node != nil {
			return true
		}
	}

	return false
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	// don't draw arrows to subnets
	if target.Type == az.SUBNET {
		return nil
	}

	// expect additional information on the Private Endpoint Azure resource to determine the resource which it points to
	peTarget := source.Properties["attachedTo"]

	// don't draw dependency arrows to the attached resource
	if target.Id == peTarget {
		return nil
	}

	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
