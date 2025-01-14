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

func (*handler) DrawIcon(resource *az.Resource, resources *map[string]*node.ResourceAndNode) []*node.Node {
	linkedResource := (*resources)[resource.Properties["attachedTo"]]

	// storage accounts might have multiple private endpoints attached to it
	if shouldExit := isOtherPrivateEndpointPointingToTheSameResource(resource, resources, linkedResource.Resource); shouldExit {
		return []*node.Node{}
	}

	return node.SetTopRightIcon(linkedResource, resources, IMAGE, HEIGHT, WIDTH)
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

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
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

	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
