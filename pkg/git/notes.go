package git

// AddNotes to HEAD in the repository.
func AddNotes(directory string, ref string, note string) error {
	args := []string{"add", "-m", note}
	_, err := Notes(directory, ref, args)
	return err
}

// ForceAddNotes to HEAD in the repository.
func ForceAddNotes(directory string, ref string, note string) error {
	args := []string{"add", "-f", "-m", note}
	_, err := Notes(directory, ref, args)
	return err
}

// AppendNotes to HEAD in the repository.
func AppendNotes(directory string, ref string, note string) error {
	args := []string{"append", "-m", note}
	_, err := Notes(directory, ref, args)
	return err
}

// ListNotes in HEAD in the repository.
// Returns the stdout if no error.
func ListNotes(directory string, ref string) ([]byte, error) {
	args := []string{"list"}
	return Notes(directory, ref, args)
}

// ShowNotes.
// Returns the stdout if there is no error.
// Defaults to HEAD if no object is given.
func ShowNotes(directory string, ref string, object string) ([]byte, error) {
	if object == "" {
		args := []string{"show"}
		return Notes(directory, ref, args)
	} else {
		args := []string{"show", object}
		return Notes(directory, ref, args)
	}
}

func RemoveNotes(directory string, ref string, object string) error {
	args := []string{"remove", object}
	_, err := Notes(directory, ref, args)
	return err
}

// Notes executes git notes on a given directory.
// Uses ref to namespace.
func Notes(directory string, ref string, cmds []string) ([]byte, error) {
	args := append([]string{"notes", "--ref", ref}, cmds...)
	return execute(directory, args)
}
