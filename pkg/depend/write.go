package depend

import "github.com/git-depend/git-depend/pkg/git"

// Write a request to a note.
func (req *Request) Write(cache *git.Cache, ref string) error {
	json, err := req.GetJson()
	if err != nil {
		return err
	}
	return cache.ForceAddNotes(req.Url, ref, string(json))
}
