package node

import (
	"azsample/internal/az"
)

type Properties struct {
	X, Y, Width, Height int
}

type ResourceAndNode struct {
	Resource *az.Resource
	Node     *Node
}
