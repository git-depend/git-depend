package git

func Fetch(directory string) error {
	args := []string{"fetch"}
	_, err := execute(directory, args)
	return err
}
