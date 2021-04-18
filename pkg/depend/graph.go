package depend

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/git-depend/git-depend/pkg/git"
	"github.com/git-depend/git-depend/pkg/utils"
)

var ref_deps_name string = "git-depend-deps"

type NodeTable map[string]*Node

// Graph node only contains its direct dependencies.
type Graph struct {
	// TODO: Perhaps implement version here.
	// Could be a sha of the dependency json.
	// version    string
	table NodeTable
	edges []*Node
	cache *git.Cache
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

// NewGraph reads in the JSON data and creates the graph.
func NewGraph(cache *git.Cache, data []byte) (*Graph, error) {
	graph := &Graph{
		cache: cache,
	}
	if err := graph.populateTableFromJson(data); err != nil {
		return nil, err
	}
	if err := graph.createGraph(); err != nil {
		return nil, err
	}
	return graph, nil
}

// Children returns a flat list of all children.
// Does not contain duplicates.
func (node *Node) Children() []*Node {
	visited := utils.NewSet()
	var children []*Node
	for _, d := range node.deps {
		if !visited.Exists(d.name) {
			visited.Add(d.name)
			children = append(children, d)
		}
		for _, c := range d.Children() {
			if !visited.Exists(c.name) {
				visited.Add(c.name)
				children = append(children, c)
			}
		}
	}
	return children
}

// URLs returns all URLs in the graph.
func (graph *Graph) URLs() []string {
	urls := utils.NewSet()
	for _, e := range graph.edges {
		urls.Add(e.url)
		for _, node := range e.Children() {
			if !urls.Exists(node.url) {
				urls.Add(node.url)
			}
		}
	}
	return urls.Iterate()
}

// populateTable will create the NodeTable.
// It does not check for cycles.
func (graph *Graph) populateTableFromJson(data []byte) error {
	var repo_list []*repo
	if err := json.Unmarshal(data, &repo_list); err != nil {
		return err
	}

	table := make(NodeTable, len(repo_list))
	// Collect a map which contains a list of the dependencies.
	deps := make(map[string][]string)
	for _, repo := range repo_list {
		// Populate a table which contains a list of dependencies.
		if _, ok := deps[repo.Name]; !ok {
			// Collect the list of dependencies.
			deps[repo.Name] = repo.Deps
			// Don't collect the dependencies yet.
			table[repo.Name] = &Node{
				name: repo.Name,
				url:  repo.URL,
			}
		} else {
			return errors.New("Duplicate key: " + repo.Name)
		}
	}

	// Populate only direct dependencies.
	for k, v := range table {
		repo_deps := deps[k]
		v.deps = make([]*Node, len(repo_deps))
		for i, d := range repo_deps {
			node, ok := table[d]
			if !ok {
				return fmt.Errorf("dependency (%s) does not exist for repository %s", d, k)
			}
			v.deps[i] = node
		}
	}
	graph.table = table
	return nil
}

// createGraph will find the edges and populate the graph.
func (graph *Graph) createGraph() error {
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

	// Populate graph edges.
	graph.edges = make([]*Node, len(edges))
	i := 0
	for _, v := range edges {
		graph.edges[i] = v
		i++
	}

	// See if there were any edges.
	if len(graph.edges) == 0 {
		return errors.New("no edges in graph")
	}

	// Find cycles.
	visited := utils.NewSet()
	for _, v := range graph.edges {
		if err := v.findNodeCycles(visited); err != nil {
			return err
		}
	}

	// Check for unreachable parts of the graph.
	if (len(visited.Iterate()) + len(graph.edges)) != len(graph.table) {
		// Todo: Improve errors.
		return errors.New("unreachable parts of the graph")
	}

	return nil
}

func (node *Node) findNodeCycles(visited *utils.StringSet) error {
	for _, dep := range node.deps {
		if !visited.Exists(dep.name) {
			visited.Add(dep.name)
			children_found, err := dep.findChildren(dep.name, visited)
			if err != nil {
				return err
			}
			for _, c := range children_found {
				if dep.name == c.name {
					return &NodeCycleError{dep.name, dep.name}
				}
			}
		}
		if err := dep.findNodeCycles(visited); err != nil {
			return err
		}
	}
	return nil
}

// findChildren is used when constructing the graph.
func (node *Node) findChildren(visiting string, visited *utils.StringSet) ([]*Node, error) {
	var children []*Node
	for _, d := range node.deps {
		if visiting == d.name {
			return nil, &NodeCycleError{visiting, node.name}
		}
		if !visited.Exists(d.name) {
			visited.Add(d.name)
			children = append(children, d)
		}
		children_found, err := d.findChildren(visiting, visited)
		if err != nil {
			return nil, err
		}
		for _, c := range children_found {
			if !visited.Exists(c.name) {
				visited.Add(c.name)
				children = append(children, c)
			}
		}
	}
	return children, nil
}

func (node *Node) dependencyNames() []string {
	deps := make([]string, len(node.deps))
	for i, d := range node.deps {
		deps[i] = d.name
	}
	return deps
}

// directChildRepos returns a flat list of repositories.
func (node *Node) directChildRepos() []*repo {
	deps := node.deps
	repos := make([]*repo, len(deps))
	for i, node := range deps {
		repos[i] = &repo{
			node.name,
			node.url,
			node.dependencyNames(),
		}
	}
	return repos
}

//PopulateDependencyNotes with dependency information.
func (node *Node) PopulateDependencyNotes(ID string, cache *git.Cache) error {
	lock := NewLock(ID, cache)
	repos := node.directChildRepos()
	log.Println(repos)
	data, err := json.MarshalIndent(repos, "", "\t")
	if err != nil {
		return err
	}
	if err := lock.writeLock(node); err != nil {
		return err
	}
	if err = cache.AddNotes(node.url, ref_deps_name, string(data)); err != nil {
		return err
	}
	if err = cache.PushNotes(node.url, ref_deps_name); err != nil {
		return err
	}
	if err := lock.removeLock(node); err != nil {
		return err
	}
	return nil
}

// WriteAllDependencyNotes to this node and the children.
func (graph *Graph) WriteAllDependencyNotes(parent *Node, ID string) error {
	lock := NewLock(ID, graph.cache)
	for _, n := range parent.Children() {
		n.writeDependencyNotes(lock)
	}
	return nil
}

// WriteDependencyNotes to only this node.
func (graph *Graph) WriteDependencyNotes(node *Node, ID string) error {
	lock := NewLock(ID, graph.cache)
	node.writeDependencyNotes(lock)
	return nil
}

// writeDependencyNotes will write the note to the repository.
func (node *Node) writeDependencyNotes(lock *lock) error {
	cache := lock.cache
	repos := node.directChildRepos()
	data, err := json.MarshalIndent(repos, "", "\t")
	if err != nil {
		return err
	}
	if err := lock.writeLock(node); err != nil {
		return err
	}
	byte_s, err := cache.ListNotes(node.url, ref_deps_name)
	if err != nil {
		return err
	}
	_, object, err := parseUniqueNote(byte_s)
	if err != nil {
		return err
	}
	if object != "" {
		if err := cache.RemoveNotes(node.url, ref_deps_name, object); err != nil {
			return err
		}
	}
	if err := cache.AddNotes(node.url, ref_deps_name, string(data)); err != nil {
		return err
	}
	if err := cache.PushNotes(node.url, ref_deps_name); err != nil {
		return err
	}
	if err := lock.removeLock(node); err != nil {
		return err
	}
	return nil
}

//readDependencyNotes from the graph and return the byte data.
func (graph *Graph) readDependencyNotes(node *Node) ([]byte, error) {
	cache := graph.cache
	byte_s, err := cache.ListNotes(node.url, ref_deps_name)
	if err != nil {
		return nil, err
	}
	_, object, err := parseUniqueNote(byte_s)
	if err != nil {
		return nil, err
	}
	byte_s, err = cache.ShowNotes(node.url, ref_deps_name, object)
	if err != nil {
		return nil, err
	}
	return byte_s, nil
}
