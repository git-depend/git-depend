package git

// Merge will use --ff-only.
func Merge(directory string, branch string) error {
	args := []string{
		"merge",
		"--ff-only",
		branch,
	}
	_, err := execute(directory, args)
	return err
}
