package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	ctlwork "github.com/k14s/kwt/pkg/kwt/workspace"
)

type RemoteInput struct {
	input    ctlwork.UploadInput
	uploadOp UploadOperation
	files    *inputFiles
}

func NewRemoteInput(input ctlwork.UploadInput, uploadOp UploadOperation) RemoteInput {
	return RemoteInput{input, uploadOp, newInputFiles()}
}

func (i RemoteInput) Sync() error {
	// TODO gap between initial upload and next syncOnce
	err := i.initFiles()
	if err != nil {
		return err
	}

	for {
		err := i.syncOnce()
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}
}

func (i RemoteInput) initFiles() error {
	return i.listFiles(func(path string, info os.FileInfo) error {
		i.files.MarkExists(path, info)
		return nil
	})
}

func (i RemoteInput) syncOnce() error {
	// t1 := time.Now()
	// defer func() {
	// 	fmt.Printf("---> took %s\n", time.Now().Sub(t1))
	// }()

	var changedPaths []string

	i.files.MarkAllNotExists()

	err := i.listFiles(func(path string, info os.FileInfo) error {
		if i.files.IsChanged(path, info) {
			changedPaths = append(changedPaths, path)
		}
		i.files.MarkExists(path, info)
		return nil
	})
	if err != nil {
		return err
	}

	deletedPaths := i.files.Deleted()

	if len(changedPaths) > 0 || len(deletedPaths) > 0 {
		for _, path := range changedPaths {
			fmt.Printf("changed: %s\n", path)
		}

		for _, path := range deletedPaths {
			fmt.Printf("deleted: %s\n", path)
		}

		return i.uploadOp.Run()
	}

	return nil
}

func (i RemoteInput) listFiles(callbackFunc func(string, os.FileInfo) error) error {
	return filepath.Walk(i.input.LocalPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		return callbackFunc(path, info)
	})
}

type inputFiles struct {
	files map[string]inputFile
}

func newInputFiles() *inputFiles {
	return &inputFiles{files: map[string]inputFile{}}
}

func (f *inputFiles) IsChanged(path string, newInfo os.FileInfo) bool {
	if prevInfo, found := f.files[path]; found {
		return newInfo.ModTime().UnixNano() != prevInfo.ModTime().UnixNano() ||
			newInfo.Size() != prevInfo.Size()
	}
	return true
}

func (f *inputFiles) MarkExists(path string, newInfo os.FileInfo) {
	f.files[path] = inputFile{FileInfo: newInfo, Exists: true}
}

func (f *inputFiles) MarkAllNotExists() {
	for path, file := range f.files {
		file.Exists = false
		f.files[path] = file
	}
}

func (f *inputFiles) Deleted() []string {
	var result []string
	for path, file := range f.files {
		if !file.Exists {
			result = append(result, path)
		}
	}
	return result
}

type inputFile struct {
	os.FileInfo
	Exists bool
}
