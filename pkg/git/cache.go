package git

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path"
	"sync"

	"github.com/git-depend/git-depend/pkg/utils"
)

// Stores the path of the cache
type Cache struct {
	sync.Mutex
	// root path to the cache
	root string
	// repositories maps the URL to the filepath
	repositories map[string]string
}

type CacheError struct {
	Errors map[string]error
}

func (e *CacheError) Error() string {
	msg := ""
	for k, v := range e.Errors {
		msg += fmt.Sprintf("URL: %s\nError: %s\n\n", k, v)
	}
	return msg
}

func NewCache(path string) (*Cache, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	}
	return &Cache{
		root:         path,
		repositories: make(map[string]string),
	}, nil
}

func (cache *Cache) getNewHasher() hash.Hash {
	return sha1.New()
}

// CloneOrUpdate figures out if we need to clone the repo or simply update.
// Clones are done by taking the shasum of the url and using that as the directory name.
// Returns the sha and error.
func (cache *Cache) CloneOrUpdate(url string) (string, error) {
	h := cache.getNewHasher()
	h.Write([]byte(url))
	sha := hex.EncodeToString(h.Sum(nil))

	directory := path.Join(cache.root, sha)
	_, err := os.Stat(directory)
	// If directory exists, add it to the map.
	// If not, create it.
	if os.IsNotExist(err) {
		// Create a tmp directory incase something goes wrong.
		tmp_directory := path.Join(cache.root, "tmp", sha)
		if err := os.RemoveAll(tmp_directory); err != nil {
			return "", err
		}
		if err := os.MkdirAll(tmp_directory, 0755); err != nil {
			return "", err
		}
		if err := Clone(url, tmp_directory); err != nil {
			return "", err
		}
		if err := os.Rename(tmp_directory, directory); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else if err := Update(url, directory); err != nil {
		return "", err
	}

	cache.Lock()
	cache.repositories[url] = sha
	cache.Unlock()

	return sha, nil
}

// CloneOrUpdateMany repositories.
// Returns the first error encountered.
func (cache *Cache) CloneOrUpdateMany(urls []string) ([]string, error) {
	var wg sync.WaitGroup
	shas_chan := make(chan struct {
		string
		error
	}, len(urls))
	set := utils.NewSet()

	for _, url := range urls {
		if exists := set.Exists(url); !exists {
			set.Add(url)
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				s, e := cache.CloneOrUpdate(url)
				shas_chan <- struct {
					string
					error
				}{s, e}
			}(url)
		}
	}
	wg.Wait()
	close(shas_chan)

	shas := make([]string, len(urls))
	errors := make(map[string]error, len(urls))
	i := 0
	for sha := range shas_chan {
		if sha.error != nil {
			errors[sha.string] = sha.error
		}
		shas[i] = sha.string
		i++
	}
	if len(errors) > 0 {
		return nil, &CacheError{errors}
	}
	return shas, nil
}

// GetRepositoryDirectory returns the directory if it exists.
// If it doesn't exist, it performs a CloneOrUpdate().
func (cache *Cache) GetRepositoryDirectory(url string) (string, error) {
	repo, ok := cache.repositories[url]
	if !ok {
		sha, err := cache.CloneOrUpdate(url)
		if err != nil {
			return "", err
		}
		return path.Join(cache.root, sha), nil
	}
	return path.Join(cache.root, repo), nil
}

// GetRepositories returns a list of the URLs.
func (cache *Cache) GetRepositories() []string {
	keys := make([]string, len(cache.repositories))
	i := 0
	for k := range cache.repositories {
		keys[i] = k
		i++
	}
	return keys
}
