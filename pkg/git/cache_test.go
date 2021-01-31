package git

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"
)

func TestCacheDir(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewCache(tempDir)

	if err != nil {
		t.Log("Unexpected error: ", err)
		t.Fail()
	}
	if tempDir != cache.root {
		t.Log("root path not correct")
		t.Log(tempDir)
		t.Log(cache.root)
		t.Fail()
	}
}

func TestCacheKey(t *testing.T) {
	urlA := createLocalGitRepo(t)
	urlB := createLocalGitRepo(t)

	tempDir := t.TempDir()
	cache, _ := NewCache(tempDir)

	if _, err := cache.CloneOrUpdate(urlA); err != nil {
		t.Fatal("Error: ", err)
	}

	if _, ok := cache.repositories[urlA]; !ok {
		t.Fatal("Cache key not found.")
	}

	if len(cache.repositories) != 1 {
		t.Fatal("Cache size wrong.")
	}

	if _, err := cache.CloneOrUpdate(urlB); err != nil {
		t.Fatal("Error: ", err)
	}

	if _, ok := cache.repositories[urlB]; !ok {
		t.Fatal("Cache key not found.")
	}

	if len(cache.repositories) != 2 {
		t.Fatal("Cache size wrong.")
	}

	if _, err := cache.CloneOrUpdate(urlA); err != nil {
		t.Fatal("Error: ", err)
	}

	if _, ok := cache.repositories[urlA]; !ok {
		t.Fatal("Cache key not found.")
	}

	if len(cache.repositories) != 2 {
		t.Fatal("Cache size wrong.")
	}
}

func TestCacheCreateDirectory(t *testing.T) {
	urlA := createLocalGitRepo(t)

	tempDir := t.TempDir()
	cache, _ := NewCache(tempDir)

	if _, err := cache.CloneOrUpdate(urlA); err != nil {
		t.Fatal("Error: ", err)
	}

	dir := path.Join(tempDir, cache.repositories[urlA])
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		t.Fatal("Folder not created: ", dir)
	}
}

func TestCacheDirectoryAlreadyExists(t *testing.T) {
	urlA := createLocalGitRepo(t)

	tempDir := t.TempDir()
	cache, _ := NewCache(tempDir)

	h := cache.getNewHasher()
	h.Write([]byte(urlA))
	sha := hex.EncodeToString(h.Sum(nil))
	folder := path.Join(tempDir, sha)

	// Create the directory with clone
	Clone(urlA, folder)

	if _, err := cache.CloneOrUpdate(urlA); err != nil {
		t.Fatal("Error: ", err)
	}

	dir := path.Join(tempDir, cache.repositories[urlA])
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		t.Fatal("Folder not created: ", dir)
	}
}

func TestCloneUpdateMany(t *testing.T) {
	n_urls := []int{1, 3, 100}

	for _, n_url := range n_urls {
		// Setup
		urls := make([]string, n_url)
		for i := 0; i < n_url; i++ {
			urls[i] = createLocalGitRepo(t)
		}

		tempDir := t.TempDir()
		cache, _ := NewCache(tempDir)

		// Run
		t.Run(strconv.Itoa(n_url), func(t *testing.T) {
			shas, err := cache.CloneOrUpdateMany(urls)
			if err != nil {
				t.Fatal("Error: ", err)
			}

			if len(shas) != n_url {
				fmt.Println("Number of shas: ", n_urls)
				fmt.Println("Expected: ", len(urls))
				t.Fatal("Incorrect number of shas returned.")
			}
		})
	}
}

func TestGetRepositoryDirectory(t *testing.T) {
	urlA := createLocalGitRepo(t)
	tempDir := t.TempDir()
	cache, _ := NewCache(tempDir)

	h := cache.getNewHasher()
	h.Write([]byte(urlA))
	sha := hex.EncodeToString(h.Sum(nil))
	folder := path.Join(tempDir, sha)

	dir, err := cache.GetRepositoryDirectory(urlA)

	if err != nil {
		t.Fatal("Error: ", err)
	}

	if folder != dir {
		fmt.Println("Incorrect directory returned.")
		t.Fatalf("%s != %s", folder, dir)
	}
}

func TestGetRepositoryDirectoryAlreadyExists(t *testing.T) {
	urlA := createLocalGitRepo(t)
	tempDir := t.TempDir()
	cache, _ := NewCache(tempDir)

	h := cache.getNewHasher()
	h.Write([]byte(urlA))
	sha := hex.EncodeToString(h.Sum(nil))
	folder := path.Join(tempDir, sha)

	Clone(urlA, folder)

	dir, err := cache.GetRepositoryDirectory(urlA)

	if err != nil {
		t.Fatal("Error: ", err)
	}

	if folder != dir {
		fmt.Println("Incorrect directory returned.")
		t.Fatalf("%s != %s", folder, dir)
	}
}

func TestGetRepositories(t *testing.T) {
	n_urls := []int{1, 3, 100}

	for _, n_url := range n_urls {
		// Setup
		urls := make([]string, n_url)
		for i := 0; i < n_url; i++ {
			urls[i] = createLocalGitRepo(t)
		}

		tempDir := t.TempDir()
		cache, _ := NewCache(tempDir)

		// Run
		t.Run(strconv.Itoa(n_url), func(t *testing.T) {
			_, err := cache.CloneOrUpdateMany(urls)
			if err != nil {
				t.Fatal("Error: ", err)
			}

			dirs := cache.GetRepositories()

			if len(dirs) != n_url {
				fmt.Println("Number of shas: ", n_urls)
				fmt.Println("Expected: ", len(urls))
				t.Fatal("Incorrect number of shas returned.")
			}
		})
	}
}

// Creates an empty git repo in a temp directory.
// Returns file://{dir}
func createLocalGitRepo(t *testing.T) string {
	dir := t.TempDir()
	cmd := exec.Command("git", "-C", dir, "init")
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	return "file://" + dir
}
