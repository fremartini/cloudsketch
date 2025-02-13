package node

import (
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

func ScaleDownAndSetIconBottomLeft(iconToMove *Node, relativeTo *Node) {
	relativeToGeometry := relativeTo.GetGeometry()
	iconToMoveGeometry := iconToMove.GetGeometry()

	iconToMove.SetDimensions(iconToMoveGeometry.Width/2, iconToMoveGeometry.Height/2)
	iconToMove.SetPosition(relativeToGeometry.X-(iconToMoveGeometry.Width/2), relativeToGeometry.Height-(iconToMoveGeometry.Height/2))
}

func FillResourcesInBox(box *Node, resourcesInGrouping []*Node, padding int) {
	if false && len(resourcesInGrouping) > 3 {
		fillResourcesInBoxSquare(box, resourcesInGrouping, padding)

		return
	}

	fillResourcesInBoxLine(box, resourcesInGrouping, padding)
}

func tallestResource(resourcesInGrouping []*Node) int {
	heightValues := list.Map(resourcesInGrouping, func(r *Node) int {
		return r.GetGeometry().Height
	})

	tallest := list.Fold(heightValues, 0, func(acc, height int) int {
		return int(math.Max(float64(acc), float64(height)))
	})

	return tallest
}

func fillResourcesInBoxLine(box *Node, resources []*Node, padding int) {
	tallestResource := tallestResource(resources)

	nextX := padding
	boxGeometry := box.GetGeometry()
	boxGeometry.Height += padding*2 + tallestResource

	for _, resourceToPlace := range resources {
		nodeToPlaceGeometry := resourceToPlace.GetGeometry()

		offsetY := boxGeometry.Height/2 - nodeToPlaceGeometry.Height/2

		resourceToPlace.SetProperty("parent", box.Id())
		resourceToPlace.ContainedIn = box
		resourceToPlace.SetPosition(nextX, offsetY)

		nextX += nodeToPlaceGeometry.Width + padding

		boxGeometry.Width += nodeToPlaceGeometry.Width + padding
	}

	boxGeometry.Width += padding
}

func fillResourcesInBoxSquare(box *Node, resourcesInGrouping []*Node, padding int) {

	// sort by volume
	sort.Slice(resourcesInGrouping, func(i, j int) bool {
		volumeA := resourcesInGrouping[i].GetGeometry().Height + resourcesInGrouping[i].GetGeometry().Width
		volumeB := resourcesInGrouping[j].GetGeometry().Height + resourcesInGrouping[j].GetGeometry().Width

		return volumeA < volumeB
	})

	// number of rows and columns is the square root of the elements in the group
	maxRowsAndColumns := int(math.Ceil(math.Sqrt(float64(len(resourcesInGrouping)))))

	currIndx := 0

	nextX := padding
	nextY := padding

	boxGeometry := box.GetGeometry()

	for row := 0; row < maxRowsAndColumns; row++ {
		for column := 0; column < maxRowsAndColumns; column++ {
			if currIndx > len(resourcesInGrouping)-1 {
				// no more elements
				break
			}

			nodeToPlace := resourcesInGrouping[currIndx]
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
	boxGeometry.Height += resourcesInGrouping[len(resourcesInGrouping)-1].GetGeometry().Height + padding
	boxGeometry.Width += padding
}
