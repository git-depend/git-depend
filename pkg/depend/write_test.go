package depend

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/git-depend/git-depend/pkg/git"
)

// TestWriteRequests clones temporary git repositories into a cache and then writes a lock to them.
func TestWriteRequests(t *testing.T) {
	urls := []string{createLocalGitRepo(t), createLocalGitRepo(t), createLocalGitRepo(t)}

	graph, err := NewGraph(createSimpleLocalGraph(t, urls[0], urls[1], urls[2]))
	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}

	cache := createLocalGitCache(t)

	requests := NewRequests(graph.table, cache)
	if err = requests.AddRequest("foo", "branch", "main", "Eric", "eric@email.com"); err != nil {
		t.Fatal("Could not add request: " + err.Error())
	}
	if err = requests.WriteLocks(); err != nil {
		t.Fatal("Could not write requests: " + err.Error())
	}

	out, err := cache.ShowNotes(urls[0], ref_lock_name)
	if err != nil {
		t.Fatal(err)
	}
	lock := &Lock{}
	json.Unmarshal(out, &lock)

	if lock.ID != "foo" {
		t.Fatal("Lock not created: " + lock.ID)
	}

	if lock.Status != Locked.String() {
		t.Fatal("Not locked: " + Locked.String())
	}
}

func TestWriteFailedRequests(t *testing.T) {
	urls := []string{createLocalGitRepo(t), createLocalGitRepo(t), createLocalGitRepo(t)}
	graph, err := NewGraph(createSimpleLocalGraph(t, urls[0], urls[1], urls[2]))

	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}

	cache := createLocalGitCache(t)
	if _, err = cache.CloneOrUpdateMany(urls); err != nil {
		t.Fatal(err)
	}

	if err = git.AddNotes(urls[0], ref_lock_name, "Some note"); err != nil {
		t.Fatal(err)
	}

	requests := NewRequests(graph.table, cache)
	if err = requests.AddRequest("foo", "branch", "main", "Eric", "eric@email.com"); err != nil {
		t.Fatal(err)
	}

	if err = requests.WriteLocks(); err == nil {
		t.Fatal("Should have failed to create lock.")
	}
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
func createSimpleLocalGraph(t *testing.T, url1 string, url2 string, url3 string) []byte {
	repo1 := repo{
		Name: "foo",
		URL:  url1,
		Deps: []string{"bar", "baz"},
	}
	repo2 := repo{
		Name: "bar",
		URL:  url2,
	}
	repo3 := repo{
		Name: "baz",
		URL:  url3,
	}

	repos := []repo{repo1, repo2, repo3}
	data, err := json.Marshal(repos)
	if err != nil {
		t.Fatal(err)
	}

	return data
}
