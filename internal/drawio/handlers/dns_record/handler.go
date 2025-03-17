package dns_record

import (
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
	"cloudsketch/internal/list"
)

type handler struct{}

const (
	TYPE   = types.DNS_RECORD
	WIDTH  = 45
	HEIGHT = 45
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

	return node.NewGeneric(map[string]interface{}{
		"style": "shadow=0;dashed=0;html=1;strokeColor=none;fillColor=#4495D1;labelPosition=center;verticalLabelPosition=bottom;verticalAlign=top;align=center;outlineConnect=0;shape=mxgraph.veeam.dns;",
		"value": resource.Name,
	}, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	target, ok := resource.Resource.Properties["target"]

	if !ok {
		return nil
	}

	// attempt to find the resource with the target IP
	var resources []*models.Resource

	for _, v := range *resource_map {
		resources = append(resources, v.Resource)
	}

	resources = list.Filter(resources, func(r *models.Resource) bool {
		return r.Type == types.NETWORK_INTERFACE
	})

	resourceWithIp := list.FirstOrDefault(resources, nil, func(nic *models.Resource) bool {
		ip, ok := nic.Properties["ip"]

		if !ok {
			return false
		}

		return ip == target
	})

	// TODO: unable to locate. NIC has been deleted

	if resourceWithIp == nil {
		// unable to find matching IP
		return nil
	}

	resource.Resource.DependsOn = append(resource.Resource.DependsOn, resource.Resource.Id)

	return nil
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{})
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
