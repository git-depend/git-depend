package git

// AddNotes to HEAD in the repository.
func (cache *Cache) AddNotes(url string, ref string, note string) error {
	dir, err := cache.GetRepositoryDirectory(url)
	if err != nil {
		return err
	}
	args := []string{"add", "-m", note}
	_, r := Notes(dir, ref, args)
	return r
}

// ForceAddNotes to HEAD in the repository.
func (cache *Cache) ForceAddNotes(url string, ref string, note string) error {
	dir, err := cache.GetRepositoryDirectory(url)
	if err != nil {
		return err
	}
	args := []string{"add", "-f", "-m", note}
	_, r := Notes(dir, ref, args)
	return r
}

// AppendNotes to HEAD in the repository.
func (cache *Cache) AppendNotes(url string, ref string, note string) error {
	dir, err := cache.GetRepositoryDirectory(url)
	if err != nil {
		return err
	}
	args := []string{"append", "-m", note}
	_, r := Notes(dir, ref, args)
	return r
}

// ListNotes in HEAD in the repository.
// Returns the stdout if no error.
func (cache *Cache) ListNotes(url string, ref string) ([]byte, error) {
	dir, err := cache.GetRepositoryDirectory(url)
	if err != nil {
		return nil, err
	}
	args := []string{"list"}
	return Notes(dir, ref, args)
}

// ShowNotes in HEAD.
// Returns the stdout if there is no error.
func (cache *Cache) ShowNotes(url string, ref string) ([]byte, error) {
	dir, err := cache.GetRepositoryDirectory(url)
	if err != nil {
		return nil, err
	}
	args := []string{"show"}
	return Notes(dir, ref, args)
}

// Notes executes git notes on a given directory.
// Uses ref to namespace.
func Notes(directory string, ref string, cmds []string) ([]byte, error) {
	args := append([]string{"-C", directory, "notes", "--ref", ref}, cmds...)
	return execute(args)
}
