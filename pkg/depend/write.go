package depend

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/git-depend/git-depend/pkg/git"
)

var ref_lock_name string = "git-depend-lock"

type lockref struct {
	// The git object that the lock is pointing to
	object string
	// The actual note object representing the lock
	noteObject string
}

// Lock allows us to safely write to a note.
type lock struct {
	ID        string    `json:"Id"`
	Timestamp time.Time `json:"Timestamp"`
	cache     *git.Cache
	lockref   lockref
}

func NewLock(ID string, cache *git.Cache) *lock {
	return &lock{
		ID:        ID,
		Timestamp: time.Now(),
		cache:     cache,
	}
}

// writeLock will lock an individual repository.
func (lock *lock) writeLock(node *Node) error {
	data, err := json.Marshal(lock)
	if err != nil {
		return err
	}

	if err = lock.cache.AddNotes(node.url, ref_lock_name, string(data)); err != nil {
		return err
	}

	if err = lock.cache.PushNotes(node.url, ref_lock_name); err != nil {
		return err
	}

	if err = lock.populateLockRef(node); err != nil {
		return err
	}
	return nil
}

func (lock *lock) removeLock(node *Node) error {
	if err := lock.populateLockRef(node); err != nil {
		return err
	}
	if err := lock.cache.RemoveNotes(node.url, ref_lock_name, lock.lockref.object); err != nil {
		return err
	}
	if err := lock.cache.PushNotes(node.url, ref_lock_name); err != nil {
		return err
	}
	return nil
}

func (lock *lock) populateLockRef(node *Node) error {
	byte_s, err := lock.cache.ListNotes(node.url, ref_lock_name)
	if err != nil {
		return err
	}
	// For Windows compatability
	noteRefs := regexp.MustCompile("\r\n|\n").Split(strings.TrimSpace(string(byte_s)), -1)
	if len(noteRefs) > 1 {
		lastRef := strings.Fields(noteRefs[len(noteRefs)-1])[1]
		// TODO: This should be managed by an appropriate logging library
		return errors.New(fmt.Sprintf("Number of locks in the repo > 1, lock used maps to git %s", lastRef))
	}

	for _, noteRef := range noteRefs {
		n := strings.Fields(noteRef)
		if len(n) != 2 {
			return errors.New(fmt.Sprintf("Unparseable git note reference. Num of fields in list != 2 (%d)", len(n)))
		}
		lock.lockref.noteObject = n[0]
		lock.lockref.object = n[1]
	}

	return nil
}
