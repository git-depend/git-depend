package depend

import (
	"testing"

	"github.com/git-depend/git-depend/pkg/git"
)

// TestWriteRequests clones temporary git repositories into a cache and then writes a lock to them.
func TestWriteRequests(t *testing.T) {
	graph, err := NewGraph(createLocalGitCache(t), createSimpleLocalGraph(t))
	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}

	urls := graph.URLs()
	cache := createLocalGitCache(t)

	requests := NewRequests(graph.table, cache)
	if err = requests.AddRequest("foo", "branch", "main", "Eric", "eric@email.com"); err != nil {
		t.Fatal("Could not add request: " + err.Error())
	}
	if err = requests.writeLocks(); err != nil {
		t.Fatal("Could not write requests: " + err.Error())
	}

	_, err = cache.ShowNotes(urls[0], ref_lock_name, "")
	if err != nil {
		t.Fatal(err)
	}
	lock := requests.lockTable[requests.nodesTable["foo"]]
	if lock.lockref.noteObject == "" {
		t.Fatal("lockref note object reference not updated")
	}
	if lock.lockref.object == "" {
		t.Fatal("lockref git object reference not updated")
	}

	if err = requests.removeLocks(); err != nil {
		t.Fatal("Could not remove locks: " + err.Error())
	}
	_, ok := requests.lockTable[requests.nodesTable["foo"]]
	if ok {
		t.Fatal("lockTable not updated with removing lock")
	}
}

func TestWriteFailedRequests(t *testing.T) {
	graph, err := NewGraph(createLocalGitCache(t), createSimpleLocalGraph(t))

	if err != nil {
		t.Fatal("Could not create graph: " + err.Error())
	}
	urls := graph.URLs()
	cache := createLocalGitCache(t)
	if _, err = cache.CloneOrUpdateMany(urls); err != nil {
		t.Fatal(err)
	}

	node := graph.table["foo"]
	if err = git.AddNotes(node.url, ref_lock_name, "Some note"); err != nil {
		t.Fatal(err)
	}

	requests := NewRequests(graph.table, cache)
	if err = requests.AddRequest("foo", "branch", "main", "Eric", "eric@email.com"); err != nil {
		t.Fatal(err)
	}

	if err = requests.writeLocks(); err == nil {
		t.Fatal("Should have failed to create lock.")
	}
	_, ok := requests.lockTable[requests.nodesTable["foo"]]
	if ok {
		t.Fatal("lockTable wrongly updated")
	}
}
