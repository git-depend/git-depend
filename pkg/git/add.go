package git

func Add(directory string, files []string) error {
	args := append([]string{"add"}, files...)
	_, err := execute(directory, args)
	return err
}
