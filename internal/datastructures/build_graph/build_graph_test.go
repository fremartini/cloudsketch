package build_graph_test

import (
	"cloudsketch/internal/datastructures/build_graph"
	"testing"
)

func TestResolveParents_ResolvesParentsFirst(t *testing.T) {
	// arrange
	resolvedInOrder := []string{}

	parent := build_graph.NewTask("parent", []string{}, []string{}, []string{}, func() {
		resolvedInOrder = append(resolvedInOrder, "parent")
	})
	child := build_graph.NewTask("child", []string{"parent"}, []string{}, []string{}, func() {
		resolvedInOrder = append(resolvedInOrder, "child")
	})

	tasks := []*build_graph.Task{
		parent,
		child,
	}

	bg, _ := build_graph.NewGraph(tasks)

	// act
	bg.ResolveTasksThatDependOnThis(child)

	// assert
	if resolvedInOrder[0] != "parent" {
		t.Errorf("expected parent to be resolved first")
	}

	if resolvedInOrder[1] != "child" {
		t.Errorf("expected child to be resolved last")
	}
}

func TestResolveChildren_ResolvesChildrenFirst(t *testing.T) {
	// arrange
	resolvedInOrder := []string{}

	parent := build_graph.NewTask("parent", []string{}, []string{}, []string{}, func() {
		resolvedInOrder = append(resolvedInOrder, "parent")
	})

	child := build_graph.NewTask("child", []string{"parent"}, []string{}, []string{}, func() {
		resolvedInOrder = append(resolvedInOrder, "child")
	})

	tasks := []*build_graph.Task{
		parent,
		child,
	}

	bg, _ := build_graph.NewGraph(tasks)

	// act
	bg.ResolveDependencies(parent)

	// assert
	if resolvedInOrder[0] != "child" {
		t.Errorf("expected child to be resolved first")
	}

	if resolvedInOrder[1] != "parent" {
		t.Errorf("expected parent to be resolved last")
	}
}
