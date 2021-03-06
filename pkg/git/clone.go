package git

// Clone also adds a fetch for notes.
func Clone(url string, directory string) error {
	args := []string{
		"clone",
		"-c",
		"remote.origin.fetch='refs/notes/*:refs/notes/*'",
		url,
		directory,
	}
	_, err := execute("", args)
	return err
}
