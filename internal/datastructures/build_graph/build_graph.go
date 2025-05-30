package build_graph

import (
	"cloudsketch/internal/list"
	"errors"
	"fmt"
	"sort"
)

type Build_graph struct {
	Tasks         []*Task
	Graph         map[string][]*Task
	Inverse_graph map[string][]*Task
}

func NewGraph(tasks []*Task) (*Build_graph, error) {
	graph, inverse_graph, err := buildGraph(tasks)

	if err != nil {
		return nil, err
	}

	return &Build_graph{
		Tasks:         tasks,
		Graph:         graph,
		Inverse_graph: inverse_graph,
	}, nil
}

type Task struct {
	Label      string
	References []string
	Inputs     []string
	Outputs    []string
	Action     func()
}

func NewTask(label string, references, inputs, outputs []string, action func()) *Task {
	return &Task{
		Label:      label,
		References: references,
		Inputs:     inputs,
		Outputs:    outputs,
		Action:     action,
	}
}

func buildGraph(tasks []*Task) (map[string][]*Task, map[string][]*Task, error) {
	graph := map[string][]*Task{}
	inverse_graph := map[string][]*Task{}

	// sort tasks lowest amount of dependencies first
	sort.Slice(tasks, func(i, j int) bool {
		return len(tasks[i].References) > len(tasks[j].References)
	})

	// if the last entry has refernces the graph is cyclic
	if len(tasks[len(tasks)-1].References) != 0 {
		return nil, nil, errors.New("cyclic graph detected")
	}

	for _, task := range tasks {
		_, ok := graph[task.Label]
		if !ok {
			graph[task.Label] = []*Task{}
		}

		for _, reference := range task.References {
			dependentTask := list.FirstOrDefault(tasks, nil, func(t *Task) bool {
				return t.Label == reference
			})

			if dependentTask == nil {
				return nil, nil, fmt.Errorf("an unknown task was referenced %s", reference)
			}

			graph[task.Label] = append(graph[task.Label], dependentTask)

			_, ok := inverse_graph[task.Label]
			if !ok {
				inverse_graph[task.Label] = []*Task{}
			}

			inverse_graph[dependentTask.Label] = append(inverse_graph[dependentTask.Label], task)
		}
	}

	return graph, inverse_graph, nil
}

func (g *Build_graph) ResolveInverse(t *Task) {
	e := g.Inverse_graph[t.Label]

	for _, ref := range e {
		// recursively resolve the tasks dependencies
		g.ResolveInverse(ref)
	}

	// when the task has no dependencies it can be resolved
	t.Action()
}

func (g *Build_graph) Resolve(t *Task) {
	e := g.Graph[t.Label]

	for _, ref := range e {
		// recursively resolve the tasks dependencies
		g.Resolve(ref)
	}

	// when the task has no dependencies it can be resolved
	t.Action()
}
