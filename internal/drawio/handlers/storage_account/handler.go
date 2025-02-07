package storage_account

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
	"azsample/internal/list"
)

type handler struct{}

const (
	TYPE   = az.STORAGE_ACCOUNT
	IMAGE  = images.STORAGE_ACCOUNT
	WIDTH  = 65
	HEIGHT = 52
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
	privateEndpoints := getPrivateEndpointPointingToResource(resource_map, resource.Resource)

	if len(privateEndpoints) <= 1 {
		return nil
	}

	resources := []*az.Resource{}
	for _, e := range *resource_map {
		resources = append(resources, e.Resource)
	}

	firstSubnet := getPrivateEndpointSubnet(privateEndpoints[0], resources)

	// if all private endpoint belong to the same subnet
	allPrivateEndpointsInSameSubnet := list.Fold(privateEndpoints[1:], true, func(resource *az.Resource, matches bool) bool {
		privateEndpointSubnet := getPrivateEndpointSubnet(resource, resources)

		return matches && privateEndpointSubnet == firstSubnet
	})

	if !allPrivateEndpointsInSameSubnet {
		return nil
	}

	// delete unneeded private endpoint icons
	for _, pe := range privateEndpoints[1:] {
		delete(*resource_map, pe.Id)
	}

	// set icon top right
	privateEndpointToMove := (*resource_map)[privateEndpoints[0].Id]
	return node.SetIcon(resource.Node, privateEndpointToMove.Node, node.TOP_RIGHT)
}

func getPrivateEndpointSubnet(resource *az.Resource, resources []*az.Resource) *string {
	for _, dependency := range resource.DependsOn {
		// TODO: refactor this!
		if c := list.Contains(resources, func(resource *az.Resource) bool {
			return resource.Id == dependency
		}); !c {
			continue
		}

		resource := list.First(resources, func(resource *az.Resource) bool {
			return resource.Id == dependency
		})

		if resource.Type == az.SUBNET {
			return &resource.Id
		}
	}

	return nil
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
	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
