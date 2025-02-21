package diagram

import (
	"bufio"
	"bytes"
	"cloudsketch/internal/guid"
	"fmt"
	"os"
)

const (
	Padding = 50
)

var (
	BoxOriginX     = 0
	MaxHeightSoFar = 0
)

type diagram struct {
	cells []string
}

func New(cells []string) *diagram {
	return &diagram{
		cells: cells,
	}
}

func (d *diagram) Write(filename string) error {
	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	var buffer bytes.Buffer

	for _, cell := range d.cells {
		buffer.WriteString(cell)
	}

	diagramId := guid.NewGuidAlphanumeric()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(fmt.Sprintf(`<mxfile host="Electron" agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) draw.io/25.0.1 Chrome/128.0.6613.186 Electron/32.2.6 Safari/537.36" version="25.0.1">
	<diagram name="Page-1" id="%s">
		<mxGraphModel dx="2074" dy="1196" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="850" pageHeight="1100" math="0" shadow="0">
			<root>
				<mxCell id="0" />
				<mxCell id="1" parent="0" />%s
			</root>
		</mxGraphModel>
	</diagram>
</mxfile>
`, diagramId, buffer.String()))

	if err != nil {
		return err
	}

	return w.Flush()
}
