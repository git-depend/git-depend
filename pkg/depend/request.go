package depend

import (
	"encoding/json"
	"errors"

	"github.com/git-depend/git-depend/pkg/git"
	"github.com/git-depend/git-depend/pkg/utils"
)

// Request holds all the information about the merge which is taking place.
type Request struct {
	Name   string `json:"Name"`
	From   string `json:"From"`
	To     string `json:"To"`
	Author string `json:"Author,omitempty"`
	Email  string `json:"Email,omitempty"`
}

type RequestsTable map[*Node]*Request
type lockTable map[*Node]*lock

// Requests contains a map of Node names to the Request.
type Requests struct {
	table      RequestsTable
	nodesTable NodeTable
	cache      *git.Cache
	lockTable  lockTable
}

// NewRequests returns a new Requests struct.
func NewRequests(table NodeTable, cache *git.Cache) *Requests {
	return &Requests{
		table:      make(RequestsTable),
		nodesTable: table,
		cache:      cache,
		lockTable:  make(lockTable),
	}
}

// String returns the json from of the struct.
func (request *Request) String() (string, error) {
	out, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// AddRequest for merging.
func (requests *Requests) AddRequest(name string, from string, to string, author string, email string) error {
	node, ok := requests.nodesTable[name]
	if !ok {
		return errors.New("Node does not exist")
	}

	if _, ok := requests.table[node]; ok {
		return errors.New("Request already exists")
	}

	requests.table[node] = &Request{
		Name:   name,
		From:   from,
		To:     to,
		Author: author,
		Email:  email,
	}
	return nil
}

// Merge the requests.
func (requests *Requests) Merge() error {
	if err := requests.writeLocks(); err != nil {
		return err
	}

	for k, v := range requests.table {
		msg, err := v.String()
		if err != nil {
			return err
		}
		if err = requests.cache.Merge(k.url, v.From, v.To, msg); err != nil {
			return err
		}
	}

	if err := requests.removeLocks(); err != nil {
		return err
	}
	return nil
}

// writeLocks for a request and the children.
func (requests *Requests) writeLocks() error {
	visited := utils.NewSet()
	for node := range requests.table {
		if !visited.Exists(node.name) {
			lock := NewLock(node.name, requests.cache)
			if err := lock.writeLock(node); err != nil {
				return err
			}
			requests.lockTable[node] = lock
			visited.Add(node.name)
			for _, d := range node.Children() {
				if !visited.Exists(d.name) {
					lock := NewLock(d.name, requests.cache)
					if err := lock.writeLock(d); err != nil {
						return err
					}
					requests.lockTable[node] = lock
					visited.Add(d.name)
				}
			}
		}
	}
	return nil
}

func (requests *Requests) removeLocks() error {
	visited := utils.NewSet()
	for node := range requests.table {
		if !visited.Exists(node.name) {
			lock, ok := requests.lockTable[node]
			if ok {
				if err := lock.removeLock(node); err != nil {
					return err
				}
			}
			delete(requests.lockTable, node)
			visited.Add(node.name)
		}
		for _, d := range node.Children() {
			if !visited.Exists(d.name) {
				lock, ok := requests.lockTable[d]
				if ok {
					if err := lock.removeLock(d); err != nil {
						return err
					}
				}
				delete(requests.lockTable, node)
				visited.Add(d.name)
			}
		}
	}
	return nil
}
