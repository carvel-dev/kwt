package workspace

import (
	"path/filepath"
)

type Asset struct {
	Name          string
	LocalDirPath  string
	RemoteDirPath string
}

type UploadInput struct {
	Asset
}

type UploadInputs []UploadInput

type DownloadOutput struct {
	Asset
}

func (i Asset) LocalPath() string {
	return i.LocalDirPath
}

func (i Asset) RemotePath(workspaceDirPath string) string {
	if len(i.RemoteDirPath) > 0 {
		if filepath.IsAbs(i.RemoteDirPath) {
			return i.RemoteDirPath
		}
		return filepath.Join(workspaceDirPath, i.RemoteDirPath)
	}

	return filepath.Join(workspaceDirPath, i.Name)
}

func (is UploadInputs) FindByName(name string) (UploadInput, bool) {
	for _, i := range is {
		if i.Name == name {
			return i, true
		}
	}
	return UploadInput{}, false
}
