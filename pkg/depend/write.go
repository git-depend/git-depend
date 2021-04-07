package depend

import (
	"encoding/json"
	"time"

	"github.com/git-depend/git-depend/pkg/git"
)

var ref_lock_name string = "git-depend-lock"

// Lock allows us to safely write to a note.
type Lock struct {
	ID        string    `json:"Id"`
	Timestamp time.Time `json:"Timestamp"`
	cache     *git.Cache
}

// WriteLock will lock an individual repository.
func (lock *Lock) WriteLock(node *Node) error {
	data, err := json.Marshal(lock)
	if err != nil {
		return err
	}

	if err = lock.cache.AddNotes(node.url, ref_lock_name, string(data)); err != nil {
		return err
	}

	return lock.cache.PushNotes(node.url, ref_lock_name)
}
