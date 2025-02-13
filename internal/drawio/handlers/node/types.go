package node

import (
	"cloudsketch/internal/az"
)

type Geometry struct {
	X, Y, Width, Height int
}

type ResourceAndNode struct {
	Resource *az.Resource
	Node     *Node
}
