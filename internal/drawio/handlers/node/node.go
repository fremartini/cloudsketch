package node

import (
	"azsample/internal/guid"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Node struct {
	id       string
	values   map[string]interface{}
	geometry *Geometry
}

func NewIcon(image, label string, geometry *Geometry) *Node {
	values := map[string]interface{}{
		"style": fmt.Sprintf("image;aspect=fixed;html=1;points=[];align=center;fontSize=12;image=%s;labelBackgroundColor=none;", image),
		"value": label,
	}

	return NewGeneric(values, geometry)
}

func NewBox(geometry *Geometry, style *string) *Node {
	values := map[string]interface{}{
		"style": "rounded=0;whiteSpace=wrap;html=1;",
	}

	if style != nil {
		values["style"] = fmt.Sprintf("%s;%s", values["style"], *style)
	}

	return NewGeneric(values, geometry)
}

func NewGroup(geometry *Geometry) *Node {
	values := map[string]interface{}{
		"value":       "",
		"style":       "group",
		"connectable": "0",
	}

	return NewGeneric(values, geometry)
}

func NewGeneric(values map[string]interface{}, geometry *Geometry) *Node {
	id := guid.NewGuidAlphanumeric()

	values["id"] = id
	values["parent"] = "1"
	values["vertex"] = "1"

	return &Node{
		id:       id,
		values:   values,
		geometry: geometry,
	}
}

func (n *Node) Id() string {
	return n.id
}

func (n *Node) SetProperty(property, value string) {
	n.values[property] = value
}

func (n *Node) SetPosition(x, y int) {
	n.geometry.X = x
	n.geometry.Y = y
}

func (n *Node) SetDimensions(width, height int) {
	n.geometry.Width = width
	n.geometry.Height = height
}

func (n *Node) GetGeometry() *Geometry {
	return n.geometry
}

func (n *Node) ToMXCell() string {
	var buffer bytes.Buffer

	for k, v := range n.values {
		j, _ := json.Marshal(v)

		buffer.WriteString(fmt.Sprintf("%s=%v ", k, string(j)))
	}

	cell := fmt.Sprintf(`<mxCell %s>
	<mxGeometry x="%v" y="%v" width="%v" height="%v" as="geometry" />
</mxCell>`, strings.TrimSpace(buffer.String()), n.geometry.X, n.geometry.Y, n.geometry.Width, n.geometry.Height)

	return cell
}

func ToMXCell(n *Node) string {
	return n.ToMXCell()
}
