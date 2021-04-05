package depend

import (
	"os"
	"path"
	"testing"
)

var test_data_simple_graph string = `
[
	{
		"Name": "foo",
		"Url": "foo.git.com",
		"Deps": ["bar", "baz"]
	},
	{
		"Name": "bar",
		"Url": "bar.git.com",
		"Deps": []
	},
	{
		"Name": "baz",
		"Url": "baz.git.com",
		"Deps": []
		}
]
`

var test_data_simple_circular_graph string = `
[
	{
		"Name": "foo",
		"Url": "foo.git.com",
		"Deps": ["bar"]
	},
	{
		"Name": "bar",
		"Url": "bar.git.com",
		"Deps": ["baz"]
	},
	{
		"Name": "baz",
		"Url": "baz.git.com",
		"Deps": ["qux"]
	},
	{
		"Name": "qux",
		"Url": "baz.git.com",
		"Deps": ["foo"]
	}
]
`

var test_data_multi_graph string = `
[
	{
		"Name": "foo",
		"Url": "foo.git.com",
		"Deps": ["bar", "baz"]
	},
	{
		"Name": "bar",
		"Url": "bar.git.com",
		"Deps": []
	},
	{
		"Name": "baz",
		"Url": "baz.git.com",
		"Deps": []
	},
	{
		"Name": "qux",
		"Url": "qux.git.com",
		"Deps": ["baz"]
	}
]
`

func TestNewTableFromFile(t *testing.T) {
	temp_file := writeJson(t, test_data_simple_graph)

	graph, err := NewGraphFromFile(temp_file)
	if err != nil {
		t.Fatal(err)
	}

	if len(graph.table) != 3 {
		t.Fatal("Incorrect number of entries: ", len(graph.table))
	}

	repo, ok := graph.table["foo"]
	if !ok {
		t.Fatal("foo does not exist.")
	}

	if repo.name != "foo" {
		t.Fatal("Incorrect name: ", repo.name)
	}

	if repo.url != "foo.git.com" {
		t.Fatal("Incorrect URL: ", repo.url)
	}

	if len(repo.deps) != 2 {
		t.Fatal("Incorrect number of dependencies: ", len(repo.deps))
	}
}

func TestCreateGraph(t *testing.T) {
	temp_file := writeJson(t, test_data_simple_graph)

	graph, err := NewGraphFromFile(temp_file)
	if err != nil {
		t.Fatal(err)
	}

	if len(graph.edges) != 1 {
		t.Fatal("Incorrect number of dependencies: ", len(graph.edges))
	}

	if graph.edges[0].name != "foo" {
		t.Fatal("Incorrect name: ", graph.edges[0].name)
	}

	if len(graph.edges[0].deps) != 2 {
		t.Fatal("Incorrect number of dependencies: ", len(graph.edges[0].deps))
	}
}

func TestCiruclar(t *testing.T) {
	temp_file := writeJson(t, test_data_simple_circular_graph)
	_, err := NewGraphFromFile(temp_file)
	if err == nil {
		t.Fatal("No cycle detected.")
	}
}

func TestMultiGraph(t *testing.T) {
	temp_file := writeJson(t, test_data_multi_graph)

	graph, err := NewGraphFromFile(temp_file)
	if err != nil {
		t.Fatal(err)
	}

	if len(graph.edges) != 2 {
		t.Fatal("Incorrect number of dependencies: ", len(graph.edges))
	}

	if graph.edges[0].name != "foo" && graph.edges[1].name != "foo" {
		t.Log(graph.edges[0])
		t.Log(graph.edges[1])
		t.Fatal("foo not declared in graph.")
	}
	if graph.edges[0].name != "qux" && graph.edges[1].name != "qux" {
		t.Log(graph.edges[0])
		t.Log(graph.edges[1])
		t.Fatal("qux not declared in graph.")
	}
	for _, v := range graph.edges {
		if v.name == "qux" {
			if v.deps[0].name != "baz" {
				t.Fatal("Expected qux to contain a baz dependency.")
			}
			if v.deps[0].url != "baz.git.com" {
				t.Fatal("Expected baz to contain a baz.git.com url.")
			}
		}
	}
}

func createSimpleGraph(t *testing.T, data string) *Graph {
	temp_file := writeJson(t, data)

	graph, err := NewGraphFromFile(temp_file)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

// Writes json data to a file and returns the file path.
func writeJson(t *testing.T, data string) string {
	temp_dir := t.TempDir()
	temp_file := path.Join(temp_dir, "test.json")

	f, err := os.Create(temp_file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(data)
	if err != nil {
		t.Fatal(err)
	}

	return temp_file
}
