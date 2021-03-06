package utils

import (
	"errors"
)

type StringSet struct {
	container map[string]struct{}
}

func NewSet() *StringSet {
	return &StringSet{
		container: make(map[string]struct{}),
	}
}

func (c *StringSet) Exists(key string) bool {
	_, exists := c.container[key]
	return exists
}

func (c *StringSet) Add(key string) {
	c.container[key] = struct{}{}
}

func (c *StringSet) Remove(key string) error {
	_, exists := c.container[key]
	if !exists {
		return errors.New("item does not exist in set")
	}
	delete(c.container, key)
	return nil
}

func (c *StringSet) Size() int {
	return len(c.container)
}

func (c *StringSet) Iterate() []string {
	slice := make([]string, 0, c.Size())
	for k := range c.container {
		slice = append(slice, k)
	}
	return slice
}
