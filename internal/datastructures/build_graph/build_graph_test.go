package build_graph_test

import (
	"cloudsketch/internal/datastructures/build_graph"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
)

func f(s string) func() {
	return func() {
		fmt.Println(s)
	}
}

func TestBuildGraph(t *testing.T) {
	tasks := []*build_graph.Task{
		build_graph.NewTask("resolve_deps", []string{"compile"}, []string{"direct_deps"}, []string{}, f("resolving deps")),
		build_graph.NewTask("code_gen", []string{"compile"}, []string{}, []string{}, f("generating code")),
		build_graph.NewTask("compile", []string{"unit_test", "package"}, []string{"sources"}, []string{}, f("compiling")),
		build_graph.NewTask("unit_test", []string{}, []string{}, []string{}, f("testing the code")),
		build_graph.NewTask("package", []string{"integration_test"}, []string{}, []string{}, f("packaging files")),
		build_graph.NewTask("asset_pipeline", []string{"integration_test"}, []string{"static_input_files"}, []string{}, f("building assets")),
		build_graph.NewTask("integration_test", []string{}, []string{}, []string{"output"}, f("running integration tests")),
	}

	graph, _ := build_graph.NewGraph(tasks)

	diagram_name := "build_graph"

	diagram := graph.ToDotFile(diagram_name)

	file, _ := os.Create(fmt.Sprintf("%s.gv", diagram_name))
	defer file.Close()
	file.WriteString(diagram)

	command := fmt.Sprintf("echo '%s' | dot -Tsvg > %s.svg", diagram, diagram_name)

	cmd := exec.Command("bash", "-c", command)
	_, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}
}
