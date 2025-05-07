package dot

import (
	"bytes"
	"cloudsketch/internal/datastructures/build_graph"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/list"
	"fmt"
	"os"
	"strings"
)

type dot struct {
}

func New() *dot {
	return &dot{}
}

func removeChars(s string) string {
	r := []string{"-", "_", "/", "."}

	for _, c := range r {
		s = strings.ReplaceAll(s, c, "")
	}

	return s
}

func (d *dot) WriteDiagram(resources []*models.Resource, filename string) error {
	tasks := list.Map(resources, func(r *models.Resource) *build_graph.Task {
		return build_graph.NewTask(r.Name, list.Map(r.DependsOn, func(r *models.Resource) string { return r.Name }), []string{}, []string{}, func() {})
	})

	tasks = list.Map(tasks, func(task *build_graph.Task) *build_graph.Task {
		return &build_graph.Task{
			Label:      removeChars(task.Label),
			References: list.Map(task.References, removeChars),
			Inputs:     list.Map(task.Inputs, removeChars),
			Outputs:    list.Map(task.Outputs, removeChars),
		}
	})

	bg, err := build_graph.NewGraph(tasks)

	if err != nil {
		return err
	}

	graphName := removeChars(strings.ReplaceAll(filename, ".dot", ""))

	// directory
	if strings.Contains("/", graphName) {
		s := strings.Split(graphName, "/")
		graphName = s[len(s)-1]
	}

	content := ToDotFile(bg, graphName)

	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(content)

	return err
}

func ToDotFile(g *build_graph.Build_graph, name string) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("digraph %s {\n", name))

	for _, task := range g.Tasks {
		writeInputNodes(&buffer, task.Label, task.Inputs)

		writeReferences(&buffer, task.Label, task.References)

		writeOutputNodes(&buffer, task.Label, task.Outputs)
	}

	buffer.WriteString("}")

	return buffer.String()
}

func writeInputNodes(buffer *bytes.Buffer, label string, inputs []string) {
	for _, input := range inputs {
		buffer.WriteString("\t")
		buffer.WriteString(fmt.Sprintf(`%s [label="%s" shape=plaintext];`, input, input))
		buffer.WriteString(fmt.Sprintf("\n\t%s -> %s;\n", input, label))
	}
}

func writeReferences(buffer *bytes.Buffer, label string, references []string) {
	for _, reference := range references {
		buffer.WriteString(fmt.Sprintf("\t%s -> %s;\n", label, reference))
	}
}

func writeOutputNodes(buffer *bytes.Buffer, label string, outputs []string) {
	for _, output := range outputs {
		buffer.WriteString("\t")
		buffer.WriteString(fmt.Sprintf(`%s [label="%s" shape=plaintext];`, output, output))
		buffer.WriteString(fmt.Sprintf("\n\t%s -> %s;\n", label, output))
	}
}
