package git

func Rebase(directory string, remote string) error {
	args := []string{
		"rebase",
		remote,
	}
	_, err := execute(directory, args)
	return err
}
