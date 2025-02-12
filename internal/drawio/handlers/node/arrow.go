package node

import (
	"azsample/internal/guid"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Arrow struct {
	values         map[string]interface{}
	source, target string
}

func NewArrow(source, target string, style *string) *Arrow {
	values := map[string]interface{}{
		"id":     guid.NewGuidAlphanumeric(),
		"source": source,
		"target": target,
		"style":  "edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;jumpStyle=arc",
		"edge":   "1",
		"parent": "1",
	}

	if style != nil {
		values["style"] = fmt.Sprintf("%s;%s", values["style"], *style)
	}

	return &Arrow{
		values: values,
		source: source,
		target: target,
	}
}

func (n *Arrow) ToMXCell() string {
	var buffer bytes.Buffer

	for k, v := range n.values {
		j, _ := json.Marshal(v)

		buffer.WriteString(fmt.Sprintf("%s=%v ", k, string(j)))
	}

	cell := fmt.Sprintf(`<mxCell %s>
					<mxGeometry relative="1" as="geometry" />
				</mxCell>`,
		strings.TrimSpace(buffer.String()))

	return cell
}
