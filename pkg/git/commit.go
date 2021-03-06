package git

// Commit simply commits with a message.
func Commit(directory string, message string) error {
	args := []string{
		"commit",
		"-m",
		message,
	}
	_, err := execute(directory, args)
	return err
}

// EmptyCommit is useful for writing a single message.
func EmptyCommit(directory string, message string) error {
	args := []string{
		"commit",
		"--allow-empty",
		"-m",
		message,
	}
	_, err := execute(directory, args)
	return err
}
