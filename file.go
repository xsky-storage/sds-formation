package formation

import (
	"io"
	"os"

	"github.com/juju/errors"
)

// OpenFileFunc defines type of os.Open
type OpenFileFunc func(string, int, ...os.FileMode) (io.ReadWriteCloser, error)

func realOpenFile(fileName string, flag int, p ...os.FileMode) (io.ReadWriteCloser, error) {
	perm := os.FileMode(0)
	if len(p) != 0 {
		perm = p[0]
	}
	return os.OpenFile(fileName, flag, perm)
}

func realMakeDir(path string, perm os.FileMode) error {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	} else if err == nil {
		return nil
	}
	if err = os.Mkdir(path, perm); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// OpenFile alias of os.File
var OpenFile = realOpenFile

// Mkdir alias of os.MkMkdir
var Mkdir = realMakeDir
