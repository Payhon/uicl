package uicl

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// WriteSource replaces one existing regular file after checking its expected
// digest. It is not a multi-file transaction or a lock against arbitrary writers.
func WriteSource(path, expectedDigest string, source []byte) error {
	st, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		return fmt.Errorf("WRITE_TARGET: expected a regular file")
	}
	old, err := ReadSource(path)
	if err != nil {
		return err
	}
	if Digest(old) != expectedDigest {
		return fmt.Errorf("CONFLICT: source changed before write")
	}
	if bytes.Equal(old, source) {
		return nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".uicl-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(st.Mode().Perm()); err == nil {
		_, err = tmp.Write(source)
	}
	if err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	current, err := ReadSource(path)
	if err != nil {
		return err
	}
	now, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !now.Mode().IsRegular() || !os.SameFile(st, now) || Digest(current) != expectedDigest {
		return fmt.Errorf("CONFLICT: source changed before replacement")
	}
	return os.Rename(name, path)
}
