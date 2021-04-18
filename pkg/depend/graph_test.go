package depend

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/git-depend/git-depend/pkg/git"
)

func TestCreateGraph(t *testing.T) {
	graph, err := NewGraph(createLocalGitCache(t), createSimpleLocalGraph(t))
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
	_, err := NewGraph(createLocalGitCache(t), createSimpleLocalCircularGraph(t))
	if err == nil {
		t.Fatal("No cycle detected.")
	}
}

func TestMultiGraph(t *testing.T) {
	graph, err := NewGraph(createLocalGitCache(t), createSimpleLocalMultiGraph(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.edges) != 2 {
		t.Fatal("Incorrect number of dependencies: ", len(graph.edges))
	}
	if !nodeContains(graph.edges, "foo") {
		t.Log(graph.edges[0])
		t.Log(graph.edges[1])
		t.Fatal("foo not declared in graph.")
	}
	if !nodeContains(graph.edges, "qux") {
		t.Log(graph.edges[0])
		t.Log(graph.edges[1])
		t.Fatal("qux not declared in graph.")
	}
	for _, v := range graph.edges {
		if v.name == "qux" {
			if v.deps[0].name != "baz" {
				t.Fatal("Expected qux to contain a baz dependency.")
			}
		}
	}
}

func TestDeepGraph(t *testing.T) {
	data := createDeepLocalGraph(t)
	graph, err := NewGraph(createLocalGitCache(t), data)
	if err != nil {
		t.Log(string(data))
		t.Fatal("Could not create graph: " + err.Error())
	}
	if len(graph.edges) != 3 {
		t.Fatalf("Wrong number of edges, wanted 3 got %d", len(graph.edges))
	}
	if !nodeContains(graph.edges, "foo") {
		t.Fatal("Expected to contain foo")
	}
	if !nodeContains(graph.edges, "qux") {
		t.Fatal("Expected to contain qux")
	}
	if !nodeContains(graph.edges, "fobble") {
		t.Fatal("Expected to contain fobble")
	}
	if len(graph.table) != 8 {
		t.Fatalf("Expected table to contain 8 nodes: %d", len(graph.table))
	}
}

func TestURLs(t *testing.T) {
	cache := createLocalGitCache(t)
	graph, err := NewGraph(cache, createDeepLocalGraph(t))
	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}
	urls := graph.URLs()
	if len(urls) != 8 {
		t.Fatalf("Expected 8 URLs: %d", len(urls))
	}
}

func TestWriteDependencyToNode(t *testing.T) {
	cache := createLocalGitCache(t)
	graph, err := NewGraph(cache, createDeepLocalGraph(t))
	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}
	if len(graph.edges) != 3 {
		t.Fatalf("Expected 3 edges: %d", len(graph.edges))
	}
	if len(graph.table) != 8 {
		t.Fatalf("Expected 8 repos: %d", len(graph.table))
	}
	node, ok := graph.table["foo"]
	if !ok {
		t.Fatalf("Graph does not contain foo.")
	}
	if err := node.PopulateDependencyNotes("id", cache); err != nil {
		t.Fatal("Could not write: " + err.Error())
	}
	data, err := cache.ShowNotes(node.url, ref_deps_name, "")
	if err != nil {
		t.Fatal(err)
	}
	var repos []*repo
	if err := json.Unmarshal(data, &repos); err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Log(string(data))
		t.Fatalf("Expected 2 dependencies: %d", len(repos))
	}
	if !repoContains(repos, "bar") {
		t.Fatal("Expected to conatin bar: " + string(data))
	}
	if !repoContains(repos, "baz") {
		t.Fatal("Expected to conatin bar: " + string(data))
	}
	if err := graph.WriteDependencyNotes(node, "id"); err != nil {
		t.Fatal("Could not write: " + err.Error())
	}
	data_read, err := graph.readDependencyNotes(node)
	if err != nil {
		t.Fatal(err)
	}
	var repos_read []*repo
	if err = json.Unmarshal(data_read, &repos_read); err != nil {
		t.Fatal(err)
	}
	if len(repos_read) != 2 {
		t.Log(string(data))
		t.Fatalf("Expected 2 dependencies: %d", len(repos))
	}
	if !repoContains(repos_read, "bar") {
		t.Fatal("Expected to conatin bar: " + string(data))
	}
	if !repoContains(repos_read, "baz") {
		t.Fatal("Expected to conatin bar: " + string(data))
	}
}

func TestWriteDependencyToNodes(t *testing.T) {
	cache := createLocalGitCache(t)
	graph, err := NewGraph(cache, createDeepLocalGraph(t))
	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}
	if len(graph.edges) != 3 {
		t.Fatalf("Expected 3 edges: %d", len(graph.edges))
	}
	if len(graph.table) != 8 {
		t.Fatalf("Expected 8 repos: %d", len(graph.table))
	}
	node, ok := graph.table["foo"]
	if !ok {
		t.Fatalf("Graph does not contain foo.")
	}
	if err := node.PopulateDependencyNotes("id", cache); err != nil {
		t.Fatal("Could not write: " + err.Error())
	}
	data, err := cache.ShowNotes(node.url, ref_deps_name, "")
	if err != nil {
		t.Fatal(err)
	}
	var repos []*repo
	if err := json.Unmarshal(data, &repos); err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Log(string(data))
		t.Fatalf("Expected 2 dependencies: %d", len(repos))
	}
	if !repoContains(repos, "bar") {
		t.Fatal("Expected to conatin bar: " + string(data))
	}
	if !repoContains(repos, "baz") {
		t.Fatal("Expected to conatin bar: " + string(data))
	}
	if err := graph.WriteAllDependencyNotes(node, "id"); err != nil {
		t.Fatal("Could not write: " + err.Error())
	}
}

func nodeContains(nodes []*Node, name string) bool {
	for _, r := range nodes {
		if r.name == name {
			return true
		}
	}
	return false
}

func repoContains(repos []*repo, name string) bool {
	for _, r := range repos {
		if r.Name == name {
			return true
		}
	}
	return false
}

// Creates a new local git cache in a temporary directory.
func createLocalGitCache(t *testing.T) *git.Cache {
	cache, err := git.NewCache(t.TempDir())
	if err != nil {
		t.Fatal("Failed to create cache: " + err.Error())
	}
	return cache
}

// Creates a git repo in a temp directory.
func createLocalGitRepo(t *testing.T) string {
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Log(string(out))
		t.Fatal("Failed to create local git repo: " + err.Error())
	}

	empty, err := os.Create(path.Join(dir, "empty.txt"))
	if err != nil {
		t.Fatal("Failed to create file: " + err.Error())
	}
	empty.Close()

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Log(string(out))
		t.Fatal("Failed to add files: " + err.Error())
	}

	cmd = exec.Command("git", "commit", "-m", "Init.")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Log(string(out))
		t.Fatal("Failed to commit files: " + err.Error())
	}

	return dir
}

// Create the JSON for a simple graph.
func createSimpleLocalGraph(t *testing.T) []byte {
	const n_urls int = 3
	urls := make([]string, n_urls)
	for i := 0; i < n_urls; i++ {
		urls[i] = createLocalGitRepo(t)
	}
	repos := []repo{
		{
			Name: "foo",
			URL:  urls[0],
			Deps: []string{"bar", "baz"},
		},
		{
			Name: "bar",
			URL:  urls[1],
		},
		{
			Name: "baz",
			URL:  urls[2],
		},
	}
	data, err := json.MarshalIndent(repos, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Create the JSON for a deeper graph.
func createDeepLocalGraph(t *testing.T) []byte {
	const n_urls int = 8
	urls := make([]string, n_urls)
	for i := 0; i < n_urls; i++ {
		urls[i] = createLocalGitRepo(t)
	}
	repos := []repo{
		{
			Name: "foo",
			URL:  urls[0],
			Deps: []string{"bar", "baz"},
		},
		{
			Name: "bar",
			URL:  urls[1],
		},
		{
			Name: "baz",
			URL:  urls[2],
			Deps: []string{"wibble", "wobble"},
		},
		{
			Name: "qux",
			URL:  urls[3],
			Deps: []string{"wobble"},
		},
		{
			Name: "wibble",
			URL:  urls[4],
			Deps: []string{"wobble"},
		},
		{
			Name: "wobble",
			URL:  urls[5],
			Deps: []string{"wubble"},
		},
		{
			Name: "wubble",
			URL:  urls[6],
		},
		{
			Name: "fobble",
			URL:  urls[7],
		},
	}
	data, err := json.MarshalIndent(repos, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Create the JSON for a simple graph with a cycle.
func createSimpleLocalCircularGraph(t *testing.T) []byte {
	const n_urls int = 4
	urls := make([]string, n_urls)
	for i := 0; i < n_urls; i++ {
		urls[i] = createLocalGitRepo(t)
	}
	repos := []repo{
		{
			Name: "foo",
			URL:  urls[0],
			Deps: []string{"bar"},
		},
		{
			Name: "bar",
			URL:  urls[1],
			Deps: []string{"baz"},
		},
		{
			Name: "baz",
			URL:  urls[2],
			Deps: []string{"qux"},
		},
		{
			Name: "qux",
			URL:  urls[3],
			Deps: []string{"bar"},
		},
	}
	data, err := json.MarshalIndent(repos, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Create the JSON for a simple graph with multiple edges
func createSimpleLocalMultiGraph(t *testing.T) []byte {
	const n_urls int = 4
	urls := make([]string, n_urls)
	for i := 0; i < n_urls; i++ {
		urls[i] = createLocalGitRepo(t)
	}
	repos := []repo{
		{
			Name: "foo",
			URL:  urls[0],
			Deps: []string{"bar", "baz"},
		},
		{
			Name: "bar",
			URL:  urls[1],
		},
		{
			Name: "baz",
			URL:  urls[2],
		},
		{
			Name: "qux",
			URL:  urls[3],
			Deps: []string{"baz"},
		},
	}
	data, err := json.MarshalIndent(repos, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	return data
}
