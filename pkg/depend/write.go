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

	note_object, object, err := parseUniqueNote(byte_s)
	if err != nil {
		return err
	}

	lock.lockref.noteObject = note_object
	lock.lockref.object = object
	return nil
}

func parseUniqueNote(note []byte) (string, string, error) {
	// For Windows compatability
	noteRefs := regexp.MustCompile("\r\n|\n").Split(strings.TrimSpace(string(note)), -1)
	if len(noteRefs) == 1 {
		n := strings.Fields(noteRefs[0])
		if len(n) != 2 {
			return "", "", fmt.Errorf("unparseable git note reference, num of fields in list != 2 (%d)", len(n))
		}
		return n[0], n[1], nil
	} else if len(noteRefs) == 0 {
		return "", "", nil
	}

	lastRef := strings.Fields(noteRefs[len(noteRefs)-1])[1]
	// TODO: This should be managed by an appropriate logging library
	return "", "", errors.New("number of dependencies in the repo > 1, only one dependency file needed " + lastRef)
}
