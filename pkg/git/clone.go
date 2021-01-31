package git

func Clone(url string, directory string) error {
	args := []string{"clone", url, directory}
	_, r := execute(args)
	return r
}
