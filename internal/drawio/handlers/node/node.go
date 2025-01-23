package node

import (
	"azsample/internal/guid"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Node struct {
	id         string
	values     map[string]interface{}
	properties *Properties
}

func NewIcon(image, label string, properties *Properties) *Node {
	values := map[string]interface{}{
		"style": fmt.Sprintf("image;aspect=fixed;html=1;points=[];align=center;fontSize=12;image=%s;labelBackgroundColor=none;", image),
		"value": label,
	}

	return NewGeneric(values, properties)
}

func NewBox(properties *Properties, style *string) *Node {
	values := map[string]interface{}{
		"style": "rounded=0;whiteSpace=wrap;html=1;",
	}

	if style != nil {
		values["style"] = fmt.Sprintf("%s;%s", values["style"], *style)
	}

	return NewGeneric(values, properties)
}

func NewGroup(properties *Properties) *Node {
	values := map[string]interface{}{
		"value":       "",
		"style":       "group",
		"connectable": "0",
	}

	return NewGeneric(values, properties)
}

func NewGeneric(values map[string]interface{}, properties *Properties) *Node {
	id := guid.NewGuidAlphanumeric()

	values["id"] = id
	values["parent"] = "1"
	values["vertex"] = "1"

	return &Node{
		id:         id,
		values:     values,
		properties: properties,
	}
}

func (n *Node) Id() string {
	return n.id
}

func (n *Node) SetProperty(property, value string) {
	n.values[property] = value
}

func (n *Node) SetPosition(x, y int) {
	n.properties.X = x
	n.properties.Y = y
}

func (n *Node) SetDimensions(width, height int) {
	n.properties.Width = width
	n.properties.Height = height
}

func (n *Node) GetProperties() *Properties {
	return n.properties
}

func (n *Node) ToMXCell() string {
	var buffer bytes.Buffer

	for k, v := range n.values {
		j, _ := json.Marshal(v)

		buffer.WriteString(fmt.Sprintf("%s=%v ", k, string(j)))
	}

	cell := fmt.Sprintf(`<mxCell %s>
	<mxGeometry x="%v" y="%v" width="%v" height="%v" as="geometry" />
</mxCell>`, strings.TrimSpace(buffer.String()), n.properties.X, n.properties.Y, n.properties.Width, n.properties.Height)

	return cell
}

func ToMXCell(n *Node) string {
	return n.ToMXCell()
}
