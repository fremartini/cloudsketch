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
	/*if len(resourcesInGrouping) > 3 {
		fillResourcesInBoxSquare(box, resourcesInGrouping, padding)

		return
	}*/

	fillResourcesInBoxLine(box, resourcesInGrouping, padding, setResourceParent)
}

func tallestNode(resourcesInGrouping []*Node) int {
	heightValues := list.Map(resourcesInGrouping, func(r *Node) int {
		return r.GetGeometry().Height
	})

	tallest := list.Fold(heightValues, 0, func(acc, height int) int {
		return int(math.Max(float64(acc), float64(height)))
	})

	return tallest
}

func fillResourcesInBoxLine(box *Node, nodes []*Node, padding int, setResourceParent bool) {
	tallestNode := tallestNode(nodes)

	nextX := padding
	boxGeometry := box.GetGeometry()
	boxGeometry.Height += padding*2 + tallestNode

	for _, node := range nodes {
		nodeToPlaceGeometry := node.GetGeometry()

		offsetY := boxGeometry.Height/2 - nodeToPlaceGeometry.Height/2

		if setResourceParent {
			node.SetProperty("parent", box.Id())
			node.ContainedIn = box
		}

		node.SetPosition(nextX, offsetY)

		nextX += nodeToPlaceGeometry.Width + padding

		boxGeometry.Width += nodeToPlaceGeometry.Width + padding
	}

	boxGeometry.Width += padding
}

func fillResourcesInBoxSquare(box *Node, nodes []*Node, padding int) {

	// sort by volume
	sort.Slice(nodes, func(i, j int) bool {
		volumeA := nodes[i].GetGeometry().Height + nodes[i].GetGeometry().Width
		volumeB := nodes[j].GetGeometry().Height + nodes[j].GetGeometry().Width

		return volumeA < volumeB
	})

	// number of rows and columns is the square root of the elements in the group
	maxRowsAndColumns := int(math.Ceil(math.Sqrt(float64(len(nodes)))))

	currIndx := 0

	nextX := padding
	nextY := padding

	boxGeometry := box.GetGeometry()

	for row := 0; row < maxRowsAndColumns; row++ {
		for column := 0; column < maxRowsAndColumns; column++ {
			if currIndx > len(nodes)-1 {
				// no more elements
				break
			}

			nodeToPlace := nodes[currIndx]
			nodeToPlaceGeometry := nodeToPlace.GetGeometry()

			if nodeToPlace.ContainedIn != nil {
				nodeToPlaceGeometry = nodeToPlace.ContainedIn.geometry
			}

			nodeToPlace.SetProperty("parent", box.Id())
			nodeToPlace.ContainedIn = box
			nodeToPlace.SetPosition(nextX, nextY)

			nextX += nodeToPlaceGeometry.Width + padding

			// width is only determined during the fist iteration
			if row == 0 {
				boxGeometry.Width += nodeToPlaceGeometry.Width + padding
			}

			currIndx++

			// last element, skip to new row
			if column == maxRowsAndColumns-1 {
				nextX = padding
				nextY += nodeToPlaceGeometry.Height + padding
				boxGeometry.Height += nodeToPlaceGeometry.Height + padding
			}
		}
	}

	// off by one error
	boxGeometry.Height += nodes[len(nodes)-1].GetGeometry().Height + padding
	boxGeometry.Width += padding
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

func DrawDependencyArrowsToTarget(source *models.Resource, targets []*models.Resource, resource_map *map[string]*ResourceAndNode, typeBlacklist []string) []*Arrow {
	// don't draw arrows to subscriptions
	typeBlacklist = append(typeBlacklist, types.SUBSCRIPTION)

	targetResources := list.Map(targets, func(target *models.Resource) *ResourceAndNode {
		return (*resource_map)[target.Id]
	})

	// remove entries from the blacklist
	targetResources = list.Filter(targetResources, func(target *ResourceAndNode) bool {
		return !list.Contains(typeBlacklist, func(t string) bool {
			return target.Resource.Type == t
		})
	})

	sourceNode := (*resource_map)[source.Id].Node

	// remove entries that are in the same group
	targetResources = list.Filter(targetResources, func(target *ResourceAndNode) bool {
		if sourceNode.ContainedIn == nil || target.Node.ContainedIn == nil {
			return true
		}

		hasSameGroup := sourceNode.GetParentOrThis() == target.Node.GetParentOrThis()

		return !hasSameGroup
	})

	arrows := list.Fold(targetResources, []*Arrow{}, func(target *ResourceAndNode, acc []*Arrow) []*Arrow {
		return append(acc, NewArrow(sourceNode.Id(), target.Node.Id(), nil))
	})

	return arrows
}
