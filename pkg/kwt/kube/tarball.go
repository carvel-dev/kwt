package kube

// Adapted from https://github.com/kubernetes/kubernetes/blob/f077d6736b968b320a4fcc9fe57a342eae86acc5/pkg/kubectl/cmd/cp.go#L392

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

type Tarball struct {
	reader io.Reader
	prefix string
}

func (t Tarball) Extract(destFile string) error {
	entrySeq := -1

	// TODO: use compression here?
	tarReader := tar.NewReader(t.reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		entrySeq++
		mode := header.FileInfo().Mode()
		outFileName := path.Join(destFile, clean(header.Name[len(t.prefix):]))
		baseName := path.Dir(outFileName)
		// fmt.Printf("%d %s  %s\n", mode, outFileName, baseName)

		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(outFileName, 0755); err != nil {
				return err
			}
			continue
		}

		// handle coping remote file into local directory
		if entrySeq == 0 && !header.FileInfo().IsDir() {
			exists, err := dirExists(outFileName)
			if err != nil {
				return err
			}
			if exists {
				outFileName = filepath.Join(outFileName, path.Base(clean(header.Name)))
			}
		}

		if mode&os.ModeSymlink != 0 {
			err := os.Symlink(header.Linkname, outFileName)
			if err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(outFileName)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	if entrySeq == -1 {
		return fmt.Errorf("no such file or directory '%s'", t.prefix) // if no file was copied
	}

	return nil
}

// clean prevents path traversals by stripping them out.
// This is adapted from https://golang.org/src/net/http/fs.go#L74
func clean(fileName string) string {
	return path.Clean(string(os.PathSeparator) + fileName)
}

// dirExists checks if a path exists and is a directory.
func dirExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
