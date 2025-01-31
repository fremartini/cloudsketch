package node

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

	// TODO: handle
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
	}

	nodeIcon.SetProperty("parent", groupId)

	return []*Node{nodeIcon, group}
}
