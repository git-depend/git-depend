package depend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/git-depend/git-depend/pkg/utils"
)

type NodeTable map[string]*Node
type repoTable map[string]*repo

// Graph node only contains its direct dependencies.
type Graph struct {
	// TODO: Perhaps implement version here.
	// Could be a sha of the dependency json.
	// version    string
	table     NodeTable
	edges     []*Node
	repoTable repoTable
}

// Node of the tree.
type Node struct {
	name string
	url  string
	deps []*Node
}

type NodeCycleError struct {
	nodeName    string
	visitedName string
}

func (e *NodeCycleError) Error() string {
	msg := fmt.Sprintf("Cycle detected.\n%s ---> %s ---> %s", e.visitedName, e.nodeName, e.visitedName)
	return msg
}

// repo contains information about the repository
// The direct dependencies in this struct are the names of other repos.
type repo struct {
	Name string   `json:"Name"`
	URL  string   `json:"Url"`
	Deps []string `json:"Deps,omitempty"`
}

// NewGraphFromFile unmarshalls JSON from a file into a graph.
func NewGraphFromFile(path string) (*Graph, error) {
	// Read and parse the data.
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewGraph(data)
}

// NewGraph reads in the JSON data and creates the graph.
func NewGraph(data []byte) (*Graph, error) {
	var reps []*repo
	if err := json.Unmarshal(data, &reps); err != nil {
		return nil, err
	}
	// Collect all the entries and check for duplications.
	repos := make(repoTable)
	for _, repo := range reps {
		// Populate the table
		if _, ok := repos[repo.Name]; !ok {
			repos[repo.Name] = repo
		} else {
			return nil, errors.New("Duplicate key: " + repo.Name)
		}
	}
	return newGraphFromRepos(repos)
}

// GetChildren returns a flat list of all children.
// Does not contain duplicates.
func (node *Node) GetChildren() []*Node {
	visited := utils.NewSet()
	var children []*Node
	for _, d := range node.deps {
		if !visited.Exists(d.name) {
			visited.Add(d.name)
			children = append(children, d)
		}
		for _, c := range d.GetChildren() {
			if !visited.Exists(c.name) {
				visited.Add(c.name)
				children = append(children, c)
			}
		}
	}
	return children
}

func newGraphFromRepos(repos repoTable) (*Graph, error) {
	graph := &Graph{
		table:     make(NodeTable, len(repos)),
		repoTable: repos,
	}

	_, err := graph.resolve(graph.repoTable, nil)
	if err != nil {
		return nil, err
	}
	graph.createGraph()
	return graph, nil
}

// resolveFromRepos will take the parsed JSON and resolve it into the Graph.Table
func (graph *Graph) resolveFromRepos(repos repoTable, visited *utils.StringSet) ([]*Node, error) {
	for k, v := range repos {
		for _, d := range v.Deps {
			if visited.Exists(d) {
				return nil, &NodeCycleError{k, d}
			}
		}
	}
	return graph.resolve(repos, visited)
}

func (graph *Graph) resolve(repos repoTable, visited *utils.StringSet) ([]*Node, error) {
	var top bool
	if visited == nil {
		top = true
	}

	var nodes []*Node
	for k, v := range repos {
		if top {
			visited = utils.NewSet()
		}
		visited.Add(k)
		node, ok := graph.table[k]
		// Check to see if we have populated this node already.
		if !ok {
			deps := make(repoTable)
			for _, d := range v.Deps {
				deps[d] = graph.repoTable[d]
			}
			nod, err := graph.resolveFromRepos(deps, visited)
			if err != nil {
				return nil, err
			}
			node = &Node{
				name: v.Name,
				url:  v.URL,
				deps: nod,
			}
			graph.table[k] = node
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// createGraph will find the edges and populate the graph.
func (graph *Graph) createGraph() {
	edges := make(NodeTable, len(graph.table))
	for k, v := range graph.table {
		edges[k] = v
	}

	// Find edges.
	for _, v := range graph.table {
		for _, d := range v.deps {
			delete(edges, d.name)
		}
	}

	// Populate graph dependencies.
	graph.edges = make([]*Node, len(edges))
	i := 0
	for _, v := range edges {
		graph.edges[i] = v
		i++
	}
}
