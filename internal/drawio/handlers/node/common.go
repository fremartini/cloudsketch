package node

func SetTopRightIcon(resource *ResourceAndNode, resources *map[string]*ResourceAndNode, icon string, height, width int) []*Node {
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

	topRightIcon := NewIcon(icon, "", &Geometry{
		X:      linkedNodeGeometry.Width - (width / 4),
		Y:      -height/2 + (height / 4),
		Width:  width / 2,
		Height: height / 2,
	})

	topRightIcon.SetProperty("parent", groupId)

	return []*Node{topRightIcon, group}
}
