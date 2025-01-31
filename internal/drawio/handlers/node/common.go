package node

import (
	"azsample/internal/list"
	"log"
	"math"
)

const (
	TOP_LEFT     = 0
	TOP_RIGHT    = 1
	BOTTOM_LEFT  = 2
	BOTTOM_RIGHT = 3
)

func SetIcon(resource *ResourceAndNode, resources *map[string]*ResourceAndNode, icon string, height, width, position int) []*Node {
	linkedNode := resource.Node
	linkedNodeGeometry := linkedNode.GetGeometry()

	// create a group on top of the referenced node, IMPORTANT: copy the geometry to avoid using the same reference
	group := NewGroup(&Geometry{
		X:      linkedNodeGeometry.X,
		Y:      linkedNodeGeometry.Y,
		Width:  linkedNodeGeometry.Width + width/4,
		Height: linkedNodeGeometry.Height,
	})
	groupId := group.Id()

	linkedNode.SetProperty("parent", groupId)
	linkedNode.SetPosition(0, 0)

	// overwrite reference to the linked resource to instead point to the group
	(*resources)[resource.Resource.Id].Node = group

	var nodeIcon *Node = nil

	w := width / 2
	y := -height/2 + (height / 4)

	switch position {
	case TOP_LEFT:
		{
			nodeIcon = NewIcon(icon, "", &Geometry{
				X:      linkedNodeGeometry.X - (width / 4),
				Y:      y,
				Width:  w,
				Height: height / 2,
			})
			break
		}
	case TOP_RIGHT:
		{
			nodeIcon = NewIcon(icon, "", &Geometry{
				X:      linkedNodeGeometry.Width - (width / 4),
				Y:      y,
				Width:  w,
				Height: height / 2,
			})
			break
		}
	case BOTTOM_LEFT:
		{
			nodeIcon = NewIcon(icon, "", &Geometry{
				X:      linkedNodeGeometry.X - (width / 4),
				Y:      y,
				Width:  w,
				Height: linkedNodeGeometry.Height + height + (height / 2),
			})
			break
		}
	case BOTTOM_RIGHT:
		{
			nodeIcon = NewIcon(icon, "", &Geometry{
				X:      linkedNodeGeometry.Width - (width / 4),
				Y:      y,
				Width:  w,
				Height: linkedNodeGeometry.Height + height + (height / 2),
			})
			break
		}
	default:
		log.Fatalf("Undefined position %v", position)
	}

	nodeIcon.SetProperty("parent", groupId)

	return []*Node{nodeIcon, group}
}

func ScaleDownAndSetIconBottomLeft(iconToMove *Node, relativeTo *Node) {
	relativeToGeometry := relativeTo.GetGeometry()
	iconToMoveGeometry := iconToMove.GetGeometry()

	iconToMove.SetDimensions(iconToMoveGeometry.Width/2, iconToMoveGeometry.Height/2)
	iconToMove.SetPosition(relativeToGeometry.X-(iconToMoveGeometry.Width/2), relativeToGeometry.Height-(iconToMoveGeometry.Height/2))
}

func FillResourcesInBox(box *Node, resourcesInGrouping []*ResourceAndNode, padding int) {
	if len(resourcesInGrouping) > 2 {
		fillResourcesInBoxSquare(box, resourcesInGrouping, padding)

		return
	}

	fillResourcesInBoxLine(box, resourcesInGrouping, padding)
}

func fillResourcesInBoxLine(box *Node, resourcesInGrouping []*ResourceAndNode, padding int) {
	// find the tallest icon among the resources
	heightValues := list.Map(resourcesInGrouping, func(r *ResourceAndNode) int {
		return r.Node.GetGeometry().Height
	})

	greatestY := list.Fold(heightValues, 0, func(acc, height int) int {
		return int(math.Max(float64(acc), float64(height)))
	})

	nextX := padding
	boxGeometry := box.GetGeometry()

	for _, resourceToPlace := range resourcesInGrouping {
		nodeToPlace := resourceToPlace.Node
		nodeToPlaceGeometry := nodeToPlace.GetGeometry()

		nodeToPlace.SetProperty("parent", box.Id())
		nodeToPlace.ContainedIn = box
		nodeToPlace.SetPosition(nextX, greatestY/2)

		nextX += nodeToPlaceGeometry.Width + padding

		boxGeometry.Width += nodeToPlaceGeometry.Width + padding
	}

	boxGeometry.Width += padding
	boxGeometry.Height = padding + greatestY
}

func fillResourcesInBoxSquare(box *Node, resourcesInGrouping []*ResourceAndNode, padding int) {
	rowsAndColumns := int(math.Ceil(math.Sqrt(float64(len(resourcesInGrouping)))))

	currIndx := 0

	nextX := padding
	nextY := padding / 2

	boxGeometry := box.GetGeometry()

	for row := 0; row < rowsAndColumns; row++ {
		for column := 0; column < rowsAndColumns; column++ {
			if currIndx > len(resourcesInGrouping)-1 {
				// no more elements
				break
			}

			resourceToPlace := resourcesInGrouping[currIndx]

			nodeToPlace := resourceToPlace.Node
			nodeToPlaceGeometry := nodeToPlace.GetGeometry()

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
			if column == rowsAndColumns-1 {
				nextX = padding
				nextY += nodeToPlaceGeometry.Height + padding/2
				boxGeometry.Height += nodeToPlaceGeometry.Height + padding
			}
		}
	}

	// off by one error
	boxGeometry.Height += resourcesInGrouping[len(resourcesInGrouping)-1].Node.GetGeometry().Height + padding/2
	boxGeometry.Width += padding
}
