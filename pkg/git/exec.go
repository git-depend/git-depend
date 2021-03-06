package git

import (
	"bytes"
	"os/exec"
)

// Offers more information that os/exec implementation.
type ExitError struct {
	Command []string
	Stderr  []byte
	Err     error
}

// Return the stderr output.
func (e *ExitError) Error() string {
	return string(e.Stderr)
}

// Simple git execution.
// Returns the stdout if there is no error.
// Error can return the stderr if there was an exec.ExitError.
func execute(directory string, command []string) ([]byte, error) {
	cmd := exec.Command("git", command...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = directory

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil, &ExitError{
				command,
				stderr.Bytes(),
				err,
			}
		}
		return nil, err
	}

	return stdout.Bytes(), nil
}
