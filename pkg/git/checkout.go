package git

// Checkout a branch.
func Checkout(directory string, branch string) error {
	args := []string{
		"checkout",
		branch,
	}
	_, err := execute(directory, args)
	return err
}

// Checkout a branch.
func CheckoutNewBranch(directory string, branch string) error {
	args := []string{
		"checkout",
		"-b",
		branch,
	}
	_, err := execute(directory, args)
	return err
}
