package node

import "cloudsketch/internal/frontends/models"

type Geometry struct {
	X, Y, Width, Height int
}

type ResourceAndNode struct {
	Resource *models.Resource
	Node     *Node
}
