package route_table

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/drawio/handlers/node"
)

type handler struct{}

const (
	TYPE   = az.ROUTE_TBLE
	WIDTH  = 40
	HEIGHT = 38
)

var (
	STYLE = "editableCssRules=.*;html=1;shape=image;verticalLabelPosition=bottom;labelBackgroundColor=none;verticalAlign=top;aspect=fixed;imageAspect=0;image=data:image/svg+xml,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnY9Imh0dHBzOi8vdmVjdGEuaW8vbmFubyIgd2lkdGg9IjIwIiBoZWlnaHQ9IjE5LjI5OTk5OTIzNzA2MDU0NyIgdmlld0JveD0iMCAwIDIwIDE5LjI5OTk5OTIzNzA2MDU0NyI+JiN4YTsJPHN0eWxlIHR5cGU9InRleHQvY3NzIj4mI3hhOwkuc3Qwe2ZpbGw6IzQyODVmNDt9JiN4YTsJLnN0MXtmaWxsOiM2NjlkZjY7fSYjeGE7CTwvc3R5bGU+JiN4YTsJPHBhdGggY2xhc3M9InN0MCIgZD0iTTIuNDMgNi4xSDBWMi42N2gzLjk0bDguNCAxMC40OWgyLjM0di0yLjcyTDIwIDE0Ljg3bC01LjMyIDQuNDN2LTIuNzFoLTMuODd6Ii8+JiN4YTsJPHBhdGggY2xhc3M9InN0MSIgZD0iTTE0LjY4IDYuMTR2Mi43MkwyMCA0LjQzIDE0LjY4IDB2Mi43MWgtMy44N0w4LjMzIDUuODJsMi4xMyAyLjY3IDEuODgtMi4zNXpNMCAxMy4ydjMuNDNoMy45NGwyLjUyLTMuMTUtMi4xMy0yLjY3LTEuOSAyLjM5eiIvPiYjeGE7PC9zdmc+;"
)

func New() *handler {
	return &handler{}
}

func (*handler) MapResource(resource *az.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	return node.NewGeneric(map[string]interface{}{
		"style": STYLE,
		"value": resource.Name,
	}, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source *az.Resource, targets []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	return []*node.Arrow{}
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
