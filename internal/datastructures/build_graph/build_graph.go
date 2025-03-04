package build_graph

import (
	"bufio"
	"bytes"
	"cloudsketch/internal/list"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
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
}

func NewTask(label string, references, inputs, outputs []string) *Task {
	return &Task{
		label:      label,
		references: references,
		inputs:     inputs,
		outputs:    outputs,
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

func (g *build_graph) ToDotFile(name string) (string, error) {
	f, err := os.Create(fmt.Sprintf("%s.gv", name))

	if err != nil {
		return "", err
	}

	defer f.Close()

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("digraph %s {\n", name))

	for _, task := range g.tasks {
		if len(task.inputs) > 0 {
			for _, input := range task.inputs {
				buffer.WriteString("\t")
				buffer.WriteString(fmt.Sprintf(`%s [label="%s" shape=plaintext];`, input, input))
				buffer.WriteString(fmt.Sprintf("\n\t%s -> %s;\n", input, task.label))
			}
		}

		if len(task.references) > 0 {
			for _, dependency := range task.references {
				buffer.WriteString(fmt.Sprintf("\t%s -> %s;\n", task.label, dependency))
			}
		}

		if len(task.outputs) > 0 {
			for _, output := range task.outputs {
				buffer.WriteString("\t")
				buffer.WriteString(fmt.Sprintf(`%s [label="%s" shape=plaintext];`, output, output))
				buffer.WriteString(fmt.Sprintf("\n\t%s -> %s;\n", task.label, output))
			}
		}
	}

	buffer.WriteString("}")

	w := bufio.NewWriter(f)
	_, err = w.WriteString(buffer.String())

	if err != nil {
		return "", err
	}

	return ReplaceMany(buffer.String(), []string{"\n", "\t"}, ""), w.Flush()
}

func ReplaceMany(s string, old []string, new string) string {
	for _, toReplace := range old {
		s = strings.ReplaceAll(s, toReplace, new)
	}

	return s
}
