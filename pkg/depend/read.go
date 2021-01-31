package depend

import "github.com/git-depend/git-depend/pkg/git"

// Read a request from a note.
func (req *Request) Read(cache *git.Cache, ref string) error {
	data, err := cache.ShowNotes(req.Url, ref)
	if err != nil {
		return err
	}
	return req.UpdateFromJson(data)
}

// Read a request from a note given a URL.
func (req *Request) ReadFromUrl(cache *git.Cache, url string, ref string) error {
	data, err := cache.ShowNotes(url, ref)
	if err != nil {
		return err
	}
	return req.UpdateFromJson(data)
}
