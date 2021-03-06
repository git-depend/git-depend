package depend

import (
	"encoding/json"
	"errors"
	"time"

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

// Requests contains a map of Node names to the Request.
type Requests struct {
	table      RequestsTable
	nodesTable NodeTable
	cache      *git.Cache
}

// NewRequests returns a new Requests struct.
func NewRequests(table NodeTable, cache *git.Cache) *Requests {
	return &Requests{
		table:      make(RequestsTable),
		nodesTable: table,
		cache:      cache,
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
	if err := requests.WriteLocks(); err != nil {
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
	return nil
}

// WriteLocks for a request and the children.
func (requests *Requests) WriteLocks() error {
	visited := utils.NewSet()
	lock := &Lock{
		ID:        "foo",
		Timestamp: time.Now(),
		Status:    Locked.String(),
		cache:     requests.cache,
	}
	for node := range requests.table {
		if !visited.Exists(node.name) {
			if err := lock.WriteLock(node); err != nil {
				return err
			}
			visited.Add(node.name)
			for _, d := range node.GetChildren() {
				if !visited.Exists(d.name) {
					if err := lock.WriteLock(d); err != nil {
						return err
					}
					visited.Add(d.name)
				}
			}
		}
	}
	return nil
}
