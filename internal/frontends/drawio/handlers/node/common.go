package node

import (
	"cloudsketch/internal/frontends/drawio/handlers/diagram"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
	"log"
	"math"
	"sort"
)

const (
	TOP_LEFT     = 0
	TOP_RIGHT    = 1
	BOTTOM_LEFT  = 2
	BOTTOM_RIGHT = 3
)

var (
	STYLE = "rounded=0;whiteSpace=wrap;html=1;dashed=1;opacity=50;"
)

func GroupIconsAndSetPosition(centerIcon, cornerIcon *Node, position int) *Node {
	centerIconGeometry := centerIcon.GetGeometry()
	cornerIconGeometry := cornerIcon.GetGeometry()

	// create a group on top of the referenced node, IMPORTANT: copy the geometry to avoid using the same reference
	group := NewGroup(&Geometry{
		X:      centerIconGeometry.X,
		Y:      centerIconGeometry.Y,
		Width:  centerIconGeometry.Width,
		Height: centerIconGeometry.Height,
	})
	groupId := group.Id()

	centerIcon.SetProperty("parent", groupId)
	centerIcon.ContainedIn = group

	widthScaled := cornerIconGeometry.Width / 2
	heightScaled := cornerIconGeometry.Height / 2

	cornerIcon.SetProperty("parent", groupId)
	cornerIcon.SetProperty("value", "")
	cornerIcon.SetDimensions(widthScaled, heightScaled)
	cornerIcon.ContainedIn = group

	SetIconRelativeTo(cornerIcon, centerIcon, position)

	return group
}

func SetIconRelativeTo(iconToMove *Node, relativeTo *Node, position int) {
	relativeToGeometry := relativeTo.GetGeometry()
	iconToMoveGeometry := iconToMove.GetGeometry()

	x := 0
	y := 0

	switch position {
	case TOP_LEFT:
		{
			x = relativeToGeometry.X
			y = relativeToGeometry.Y
			break
		}
	case TOP_RIGHT:
		{
			x = relativeToGeometry.Width
			y = relativeToGeometry.Y
			break
		}
	case BOTTOM_LEFT:
		{
			x = relativeToGeometry.X
			y = relativeToGeometry.Height
			break
		}
	case BOTTOM_RIGHT:
		{
			x = relativeToGeometry.Width
			y = relativeToGeometry.Height
			break
		}
	default:
		log.Fatalf("Undefined position %v", position)
	}

	iconToMove.SetPosition(x-iconToMoveGeometry.Width/2, y-iconToMoveGeometry.Height/2)
}

func FillResourcesInBox(box *Node, resourcesInGrouping []*Node, padding int, setResourceParent bool) {
	// sort by volume
	sort.Slice(resourcesInGrouping, func(i, j int) bool {
		volumeA := resourcesInGrouping[i].GetGeometry().Height + resourcesInGrouping[i].GetGeometry().Width
		volumeB := resourcesInGrouping[j].GetGeometry().Height + resourcesInGrouping[j].GetGeometry().Width

		return volumeA > volumeB
	})

	// the number of rows and columns is the square root of the number elements in the group
	numRowsAndColumns := int(math.Ceil(math.Sqrt(float64(len(resourcesInGrouping)))))

	startX := padding

	nextX := startX
	nextY := padding

	boxGeometry := box.GetGeometry()

	nextIdx := 0
	for column := range numRowsAndColumns {
		columnHeight := 0

		canContinueOnCurrentRow := false

		for row := 0; row < numRowsAndColumns || canContinueOnCurrentRow; row++ {
			if nextIdx >= len(resourcesInGrouping) {
				break
			}

			node := resourcesInGrouping[nextIdx]

			nodeToPlaceGeometry := node.GetGeometry()

			if setResourceParent {
				node.SetProperty("parent", box.Id())
				node.ContainedIn = box
			}

			node.SetPosition(nextX, nextY)

			nextX += nodeToPlaceGeometry.Width + padding
			boxGeometry.Width = maxInt32(nextX, boxGeometry.Width)

			columnHeight = maxInt32(nodeToPlaceGeometry.Height+padding, columnHeight)

			// on the second pass the container has been filled with the widest boxes.
			// it is now possible to arrange nodes up to the max length instead of
			// automatically moving on to a new row
			canContinueOnCurrentRow = column > 0 && nextX < boxGeometry.Width

			nextIdx++
		}

		nextX = startX

		nextY += columnHeight
		boxGeometry.Height += columnHeight
	}

	boxGeometry.Height += padding
}

func maxInt32(x, y int) int {
	return int(math.Max(float64(x), float64(y)))
}

func GetChildResourcesOfType(resources []*models.Resource, parentId, childType string, resource_map *map[string]*ResourceAndNode) []*ResourceAndNode {
	return getChildResources(resources, parentId, &childType, resource_map)
}

func GetChildResources(resources []*models.Resource, parentId string, resource_map *map[string]*ResourceAndNode) []*ResourceAndNode {
	return getChildResources(resources, parentId, nil, resource_map)
}

func getChildResources(resources []*models.Resource, parentId string, childType *string, resource_map *map[string]*ResourceAndNode) []*ResourceAndNode {
	azResources := list.Filter(resources, func(resource *models.Resource) bool {
		return list.Contains(resource.DependsOn, func(dependency *models.Resource) bool {
			return dependency.Id == parentId
		})
	})

	if childType != nil {
		azResources = list.Filter(azResources, func(r *models.Resource) bool {
			return r.Type == *childType
		})
	}

	childResources := list.Map(azResources, func(resource *models.Resource) *ResourceAndNode {
		return (*resource_map)[resource.Id]
	})

	return childResources
}

func BoxResources(parent *Node, children []*ResourceAndNode) *Node {
	parentGeometry := parent.GetGeometry()

	box := NewBox(&Geometry{
		X:      parentGeometry.X,
		Y:      parentGeometry.Y,
		Width:  0,
		Height: 0,
	}, &STYLE)

	parent.SetProperty("parent", box.Id())
	parent.ContainedIn = box
	parent.SetPosition(0, 0)

	nodesToMove := list.Map(children, func(r *ResourceAndNode) *Node {
		return r.Node.GetParentOrThis()
	})

	// move all children into the box
	FillResourcesInBox(box, nodesToMove, diagram.Padding, true)

	parent.SetDimensions(parentGeometry.Width/2, parentGeometry.Height/2)
	SetIconRelativeTo(parent, box, BOTTOM_LEFT)

	return box
}

func DrawDependencyArrowsToTargets(source *models.Resource, targets []*models.Resource, resource_map *map[string]*ResourceAndNode, typeBlacklist []string) []*Arrow {
	// don't draw arrows to subscriptions
	typeBlacklist = append(typeBlacklist, types.SUBSCRIPTION, types.VIRTUAL_NETWORK, types.SUBNET)

	// remove entries from the blacklist
	targets = list.Filter(targets, func(target *models.Resource) bool {
		return !list.Contains(typeBlacklist, func(t string) bool {
			return target.Type == t
		})
	})

	targetResources := list.Map(targets, func(target *models.Resource) *ResourceAndNode {
		return (*resource_map)[target.Id]
	})

	sourceNode := (*resource_map)[source.Id].Node

	arrows := list.Fold(targetResources, []*Arrow{}, func(target *ResourceAndNode, acc []*Arrow) []*Arrow {
		return append(acc, NewArrow(sourceNode.Id(), target.Node.Id(), nil))
	})

	return arrows
}

func HandlePrivateEndpoint(resource *ResourceAndNode, resource_map *map[string]*ResourceAndNode) *Node {
	privateEndpoints := getPrivateEndpointPointingToResource(resource_map, resource.Resource)

	if len(privateEndpoints) == 0 {
		return nil
	}

	if len(privateEndpoints) > 1 {
		// multiple private endpoints point to this resource. If they all
		// belong to the same subnet they can be merged
		resources := []*models.Resource{}
		for _, e := range *resource_map {
			resources = append(resources, e.Resource)
		}

		firstSubnet := getPrivateEndpointSubnet(privateEndpoints[0].Resource, resources)

		allPrivateEndpointsInSameSubnet := list.Fold(privateEndpoints, true, func(resource *ResourceAndNode, matches bool) bool {
			privateEndpointSubnet := getPrivateEndpointSubnet(resource.Resource, resources)

			return matches && privateEndpointSubnet == firstSubnet
		})

		if !allPrivateEndpointsInSameSubnet {
			return nil
		}

		// delete unneeded private endpoint icons
		for _, pe := range privateEndpoints {
			if pe.Resource.Id == privateEndpoints[0].Resource.Id {
				continue
			}

			delete(*resource_map, pe.Resource.Id)
		}
	}

	// one private endpoint exists, "merge" the two icons
	return GroupIconsAndSetPosition(resource.Node, privateEndpoints[0].Node, TOP_RIGHT)
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

func getPrivateEndpointPointingToResource(resource_map *map[string]*ResourceAndNode, attachedResource *models.Resource) []*ResourceAndNode {
	privateEndpoints := []*ResourceAndNode{}

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
			privateEndpoints = append(privateEndpoints, v)
		}
	}

	return privateEndpoints
}
