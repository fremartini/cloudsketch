package node

import (
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
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
	/*	if len(resourcesInGrouping) < 3 {
			fillResourcesInBoxLine(box, resourcesInGrouping, padding, setResourceParent)

			return
		}
	*/
	fillResourcesInBoxSquare(box, resourcesInGrouping, padding, setResourceParent)
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

func fillResourcesInBoxSquare(box *Node, nodes []*Node, padding int, setResourceParent bool) {
	// sort by volume
	sort.Slice(nodes, func(i, j int) bool {
		volumeA := nodes[i].GetGeometry().Height + nodes[i].GetGeometry().Width
		volumeB := nodes[j].GetGeometry().Height + nodes[j].GetGeometry().Width

		return volumeA < volumeB
	})

	padding = 0

	// number of rows and columns is the square root of the elements in the group
	numRowsAndColumns := int(math.Ceil(math.Sqrt(float64(len(nodes)))))

	startX := padding

	nextX := startX
	nextY := padding

	boxGeometry := box.GetGeometry()

	currentResourceIndex := -1
	for row := 0; row < numRowsAndColumns; row++ {
		tallestNodeThisRow := 0

		for column := 0; column < numRowsAndColumns; column++ {
			currentResourceIndex++

			if currentResourceIndex > len(nodes)-1 {
				nextY += tallestNodeThisRow + padding
				boxGeometry.Height += nextY

				// no more elements
				return
			}

			nodeToPlace := nodes[currentResourceIndex]
			nodeToPlaceGeometry := nodeToPlace.GetGeometry()

			if nodeToPlace.ContainedIn != nil {
				nodeToPlaceGeometry = nodeToPlace.ContainedIn.geometry
			}

			if setResourceParent {
				nodeToPlace.SetProperty("parent", box.Id())
				nodeToPlace.ContainedIn = box
			}

			nodeToPlace.SetPosition(nextX, nextY)

			nextX += nodeToPlaceGeometry.Width + padding

			boxGeometry.Width = maxInt32(boxGeometry.Width, nextX)

			// last element, skip to new row
			if column == numRowsAndColumns-1 {
				nextX = startX
			}

			tallestNodeThisRow = maxInt32(tallestNodeThisRow, nodeToPlaceGeometry.Height)
		}

		nextY += tallestNodeThisRow + padding
		boxGeometry.Height += nextY
	}
}

func tallestNode(resourcesInGrouping []*Node) int {
	heightValues := list.Map(resourcesInGrouping, func(r *Node) int {
		return r.GetGeometry().Height
	})

	tallest := list.Fold(heightValues, 0, func(acc, height int) int {
		return maxInt32(acc, height)
	})

	return tallest
}

func maxInt32(x, y int) int {
	return int(math.Max(float64(x), float64(y)))
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
