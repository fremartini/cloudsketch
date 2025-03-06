package build_graph

import (
	"bytes"
	"cloudsketch/internal/list"
	"errors"
	"fmt"
	"sort"
)

type build_graph struct {
	tasks         []*Task
	graph         map[*Task][]*Task
	inverse_graph map[*Task][]*Task
}

func NewGraph(tasks []*Task) (*build_graph, error) {
	graph, inverse_graph, err := buildGraph(tasks)

	if err != nil {
		return nil, err
	}

	return &build_graph{
		tasks:         tasks,
		graph:         graph,
		inverse_graph: inverse_graph,
	}, nil
}

type Task struct {
	label      string
	references []string
	inputs     []string
	outputs    []string
	action     func()
}

func NewTask(label string, references, inputs, outputs []string, action func()) *Task {
	return &Task{
		label:      label,
		references: references,
		inputs:     inputs,
		outputs:    outputs,
		action:     action,
	}
}

func buildGraph(tasks []*Task) (map[*Task][]*Task, map[*Task][]*Task, error) {
	graph := map[*Task][]*Task{}
	inverse_graph := map[*Task][]*Task{}

	// sort tasks lowest amount of dependencies first
	sort.Slice(tasks, func(i, j int) bool {
		return len(tasks[i].references) > len(tasks[j].references)
	})

	// if the last entry has refernces the graph is cyclic
	if len(tasks[len(tasks)-1].references) != 0 {
		return nil, nil, errors.New("cyclic graph detected")
	}

	for _, task := range tasks {
		_, ok := graph[task]
		if !ok {
			graph[task] = []*Task{}
		}

		for _, reference := range task.references {
			dependentTask := list.First(tasks, func(t *Task) bool {
				return t.label == reference
			})

			graph[task] = append(graph[task], dependentTask)

			_, ok := inverse_graph[task]
			if !ok {
				inverse_graph[task] = []*Task{}
			}

			inverse_graph[dependentTask] = append(inverse_graph[dependentTask], task)
		}
	}

	return graph, inverse_graph, nil
}

func (g *build_graph) Resolve(t *Task) {
	for _, ref := range g.inverse_graph[t] {
		// recursively resolve the tasks dependencies
		g.Resolve(ref)
	}

	// when the task has no dependencies it can be resolved
	t.action()
}

func (g *build_graph) ToDotFile(name string) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("digraph %s {\n", name))

	for _, task := range g.tasks {
		writeInputNodes(&buffer, task.label, task.inputs)

		writeReferences(&buffer, task.label, task.references)

		writeOutputNodes(&buffer, task.label, task.outputs)
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
