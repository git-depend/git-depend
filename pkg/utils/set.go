package utils

import "fmt"

type stringSet struct {
	container map[string]struct{}
}

func NewSet() *stringSet {
	return &stringSet{
		container: make(map[string]struct{}),
	}
}

func (c *stringSet) Exists(key string) bool {
	_, exists := c.container[key]
	return exists
}

func (c *stringSet) Add(key string) {
	c.container[key] = struct{}{}
}

func (c *stringSet) Remove(key string) error {
	_, exists := c.container[key]
	if !exists {
		return fmt.Errorf("Item doesn't exist in set.")
	}
	delete(c.container, key)
	return nil
}

func (c *stringSet) Size() int {
	return len(c.container)
}
