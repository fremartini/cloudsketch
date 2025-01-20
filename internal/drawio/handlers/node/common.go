package node

func SetTopRightIcon(resource *ResourceAndNode, resources *map[string]*ResourceAndNode, icon string, height, width int) []*Node {
	linkedNode := resource.Node
	linkedNodeProperties := linkedNode.GetProperties()

	// create a group on top of the referenced node, IMPORTANT: copy the properties to avoid using the same reference
	group := NewGroup(&Properties{
		X:      linkedNodeProperties.X,
		Y:      linkedNodeProperties.Y,
		Width:  linkedNodeProperties.Width + width/4,
		Height: linkedNodeProperties.Height,
	})
	groupId := group.Id()

	linkedNode.SetProperty("parent", groupId)
	linkedNode.SetPosition(0, 0)

	// overwrite reference to the linked resource to instead point to the group
	(*resources)[resource.Resource.Id].Node = group

	topRightIcon := NewIcon(icon, "", &Properties{
		X:      linkedNodeProperties.Width - (width / 4),
		Y:      -height/2 + (height / 4),
		Width:  width / 2,
		Height: height / 2,
	})

	topRightIcon.SetProperty("parent", groupId)

	return []*Node{topRightIcon, group}
}
