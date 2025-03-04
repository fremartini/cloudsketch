package build_graph_test

import (
	"cloudsketch/internal/datastructures/build_graph"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"
)

func TestBuildGraph(t *testing.T) {
	graph, _ := build_graph.NewGraph([]*build_graph.Task{
		build_graph.NewTask("resolve_deps", []string{"compile"}, []string{"direct_deps"}, []string{}),
		build_graph.NewTask("code_gen", []string{"compile"}, []string{}, []string{}),
		build_graph.NewTask("compile", []string{"unit_test", "package"}, []string{"sources"}, []string{}),
		build_graph.NewTask("unit_test", []string{}, []string{}, []string{}),
		build_graph.NewTask("package", []string{"integration_test"}, []string{}, []string{}),
		build_graph.NewTask("asset_pipeline", []string{"integration_test"}, []string{"static_input_files"}, []string{}),
		build_graph.NewTask("integration_test", []string{}, []string{}, []string{"output"}),
	})

	diagram_name := "build_graph"

	diagram, err := graph.ToDotFile(diagram_name)

	if err != nil {
		log.Fatal(err)
	}

	diagram = strings.ReplaceAll(diagram, fmt.Sprintf(" %s", diagram_name), "")

	command := fmt.Sprintf("echo '%s' | dot -Tsvg > %s.svg", diagram, diagram_name)

	cmd := exec.Command("bash", "-c", command)
	_, err = cmd.Output()

	if err != nil {
		log.Fatal(err)
	}
}
