package git

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
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
		t.Log("Incorrect directory returned.")
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
		t.Log("Incorrect directory returned.")
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
				t.Log("Number of shas: ", n_urls)
				t.Log("Expected: ", len(urls))
				t.Fatal("Incorrect number of shas returned.")
			}
		})
	}
}

func TestPushNotes(t *testing.T) {
	ref_lock_name := "test-lock"
	first_note := "first note"
	urls := []string{createLocalGitRepo(t), createLocalGitRepo(t), createLocalGitRepo(t)}

	cache := createLocalGitCache(t)
	err := cache.AddNotes(urls[0], ref_lock_name, first_note)
	if err != nil {
		t.Fatal(err)
	}
	err = cache.PushNotes(urls[0], ref_lock_name)
	if err != nil {
		t.Fatal(err)
	}

	out, err := cache.ShowNotes(urls[0], ref_lock_name)
	if err != nil {
		t.Fatal(err)
	}

	// Git adds a newline, so we trim it.
	// This shouldn't generally be a problem as we mostly read/write JSON.
	if strings.TrimSuffix(string(out), "\n") != (first_note) {
		t.Fatal("Could not show first note: " + string(out) + "-" + first_note)
	}
}

func TestAppendNotes(t *testing.T) {
	ref_lock_name := "test-lock"
	first_note := "first note"
	second_note := "second note"
	urls := []string{createLocalGitRepo(t), createLocalGitRepo(t), createLocalGitRepo(t)}

	cache := createLocalGitCache(t)
	err := cache.AddNotes(urls[0], ref_lock_name, first_note)
	if err != nil {
		t.Fatal(err)
	}
	err = cache.PushNotes(urls[0], ref_lock_name)
	if err != nil {
		t.Fatal(err)
	}

	out, err := cache.ShowNotes(urls[0], ref_lock_name)
	if err != nil {
		t.Fatal(err)
	}

	// Git adds a newline, so we trim it.
	// This shouldn't generally be a problem as we mostly read/write JSON.
	if strings.TrimSuffix(string(out), "\n") != (first_note) {
		t.Fatal("Could not show first note: " + string(out) + "-" + first_note)
	}

	err = cache.AppendNotes(urls[0], ref_lock_name, second_note)
	if err != nil {
		t.Fatal(err)
	}

	out, err = cache.ShowNotes(urls[0], ref_lock_name)
	if err != nil {
		t.Fatal(err)
	}

	// Git adds a newline, so we remove.
	// This shouldn't generally be a problem as we mostly read/write JSON.
	output := strings.ReplaceAll(string(out), "\n", "")
	if output != (first_note + second_note) {
		t.Fatal("Could not show first note: " + string(output) + "-" + first_note)
	}
}

// TestPushNotesFail will create two caches and try to push from both.
// This test should fail as b
func TestPushNotesRejected(t *testing.T) {
	ref_lock_name := "test-lock"
	first_note := "first note"
	second_note := "second note"

	urls := []string{createLocalGitRepo(t), createLocalGitRepo(t), createLocalGitRepo(t)}

	cache := createLocalGitCache(t)

	cache_other := createLocalGitCache(t)
	cache_other.AddNotes(urls[0], ref_lock_name, first_note)
	cache_other.PushNotes(urls[0], ref_lock_name)

	err := cache.AddNotes(urls[0], ref_lock_name, second_note)
	if err != nil {
		t.Fatal(err)
	}
	err = cache.PushNotes(urls[0], ref_lock_name)
	if err == nil {
		t.Fatal("Should not be able to push notes.")
	}
}

// Creates a new local git cache in a temporary directory.
func createLocalGitCache(t *testing.T) *Cache {
	cache, err := NewCache(t.TempDir())
	if err != nil {
		t.Fatal("Failed to create cache: " + err.Error())
	}
	return cache
}

// Creates a git repo in a temp directory.
// Returns file://{dir}
func createLocalGitRepo(t *testing.T) string {
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		t.Fatal("Failed to create local git repo: " + err.Error())
	}

	emptyFile, err := os.Create(path.Join(dir, "emptyFile.txt"))
	if err != nil {
		t.Fatal("Failed to create file: " + err.Error())
	}
	emptyFile.Close()

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = dir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Fatal("Failed to add files: " + err.Error())
	}

	cmd = exec.Command("git", "commit", "-m", "Init.")
	cmd.Dir = dir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Fatal("Failed to commit files: " + err.Error())
	}

	return "file://" + dir
}
