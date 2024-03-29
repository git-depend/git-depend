package depend

import (
	"os"
	"path"
	"testing"

	"github.com/git-depend/git-depend/pkg/git"
)

func TestNewRequest(t *testing.T) {
	graph, err := NewGraph(createLocalGitCache(t), createSimpleLocalGraph(t))
	if err != nil {
		t.Fatal(err)
	}
	reqs := NewRequests(graph.table, nil)

	if err := reqs.AddRequest("foo", "branch", "main", "Test", "test@test.com"); err != nil {
		t.Fatal("Could not add a request: " + err.Error())
	}
}

func TestNewRequestFail(t *testing.T) {
	graph, err := NewGraph(createLocalGitCache(t), createSimpleLocalGraph(t))
	if err != nil {
		t.Fatal(err)
	}
	reqs := NewRequests(graph.table, nil)

	if err := reqs.AddRequest("no-key", "branch", "main", "Test", "test@test.com"); err == nil {
		t.Fatal("Should not be able to add a new request.")
	}
}

func TestMergeRequests(t *testing.T) {
	staging_branch := "staging"
	graph, err := NewGraph(nil, createSimpleLocalGraph(t))
	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}
	node := graph.table["foo"]
	writeBranchLocalGitRepo(node.url, "other.txt", staging_branch)

	cache := createLocalGitCache(t)

	requests := NewRequests(graph.table, cache)
	if err = requests.AddRequest("foo", staging_branch, "master", "Eric", "eric@email.com"); err != nil {
		t.Fatal("Could not add request: " + err.Error())
	}
	if err = requests.Merge(); err != nil {
		t.Fatal(err)
	}

	if len(requests.lockTable) != 0 {
		t.Fatal("locks not released at end of merge")
	}
}

func writeBranchLocalGitRepo(git_path string, file_name string, branch string) error {
	file, err := os.Create(path.Join(git_path, file_name))
	if err != nil {
		return err
	}
	file.Close()

	if err = git.CheckoutNewBranch(git_path, branch); err != nil {
		return err
	}

	if err = git.Add(git_path, []string{file_name}); err != nil {
		return err
	}

	if err = git.Commit(git_path, file_name); err != nil {
		return err
	}
	return nil
}
