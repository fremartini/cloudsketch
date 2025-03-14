package node

import (
	"cloudsketch/internal/drawio/models"
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

func SetIcon(centerIcon, attachedIcon *Node, position int) *Node {
	centerIconGeometry := centerIcon.GetGeometry()
	attachedIconGeometry := attachedIcon.GetGeometry()

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

	widthScaled := attachedIconGeometry.Width / 2
	heightScaled := attachedIconGeometry.Height / 2

	attachedIcon.SetProperty("parent", groupId)
	attachedIcon.SetProperty("value", "")
	attachedIcon.SetDimensions(widthScaled, heightScaled)
	attachedIcon.ContainedIn = group

	x := 0
	y := 0

	switch position {
	case TOP_LEFT:
		{
			x = group.GetGeometry().X
			y = group.GetGeometry().Y
			break
		}
	case TOP_RIGHT:
		{
			x = group.GetGeometry().Width
			y = group.GetGeometry().Y
			break
		}
	case BOTTOM_LEFT:
		{
			x = group.GetGeometry().X
			y = group.GetGeometry().Height
			break
		}
	case BOTTOM_RIGHT:
		{
			x = group.GetGeometry().Width
			y = group.GetGeometry().Height
			break
		}
	default:
		log.Fatalf("Undefined position %v", position)
	}

	attachedIcon.SetPosition(x-widthScaled/2, y-heightScaled/2)

	return group
}

func ScaleDownAndSetIconRelativeTo(iconToMove *Node, relativeTo *Node, position int) {
	relativeToGeometry := relativeTo.GetGeometry()
	iconToMoveGeometry := iconToMove.GetGeometry()

	widthScaled := (iconToMoveGeometry.Width / 2)
	heightScaled := (iconToMoveGeometry.Height / 2)

	iconToMove.SetDimensions(widthScaled, heightScaled)

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

	iconToMove.SetPosition(x-widthScaled/2, y-heightScaled/2)
}

func FillResourcesInBox(box *Node, resourcesInGrouping []*Node, padding int) {
	if false && len(resourcesInGrouping) > 3 {
		fillResourcesInBoxSquare(box, resourcesInGrouping, padding)

		return
	}

	fillResourcesInBoxLine(box, resourcesInGrouping, padding)
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

func fillResourcesInBoxLine(box *Node, nodes []*Node, padding int) {
	tallestNode := tallestNode(nodes)

	nextX := padding
	boxGeometry := box.GetGeometry()
	boxGeometry.Height += padding*2 + tallestNode

	for _, node := range nodes {
		nodeToPlaceGeometry := node.GetGeometry()

		offsetY := boxGeometry.Height/2 - nodeToPlaceGeometry.Height/2

		node.SetProperty("parent", box.Id())
		node.ContainedIn = box
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

func DrawDependencyArrowsToTarget(source *models.Resource, targets []*models.Resource, resource_map *map[string]*ResourceAndNode, typeBlacklist []string) []*Arrow {
	arrows := []*Arrow{}

	sourceNode := (*resource_map)[source.Id].Node

	for _, target := range targets {
		target := (*resource_map)[target.Id]

		if list.Contains(typeBlacklist, func(t string) bool {
			return target.Resource.Type == t
		}) {
			continue
		}

		// if they are in the same group, don't draw the arrow
		if sourceNode.ContainedIn != nil && target.Node.ContainedIn != nil {
			if sourceNode.GetParentOrThis() == target.Node.GetParentOrThis() {
				continue
			}
		}

		arrows = append(arrows, NewArrow(sourceNode.Id(), target.Node.Id(), nil))
	}

	return arrows
}
