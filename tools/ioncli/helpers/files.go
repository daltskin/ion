package helpers

import (
	"io"
	"os"
	"path/filepath"
)

//copy a file from a source path to a destination path
func copy(src, dstFolder, dstName string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck

	out, err := os.Create(filepath.Join(dstFolder, dstName))
	if err != nil {
		return err
	}
	defer out.Close() //nolint:errcheck

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close() //nolint:errcheck
}
