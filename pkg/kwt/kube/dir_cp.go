package kube

import (
	"io"
	"path"
	"strings"
)

type DirCp struct {
	exec Exec
}

func NewDirCp(exec Exec) DirCp {
	return DirCp{exec}
}

func (s DirCp) Up(localDir, remoteDir string) error {
	reader, writer := io.Pipe()
	tarWritingErrCh := make(chan error)

	go func() {
		err := TarBuilder{}.Build(localDir, "/", TarBuilderOpts{ExcludedPaths: []string{".git"}}, writer)
		writer.Close()
		tarWritingErrCh <- err
	}()

	cmd := []string{"tar", "xf", "-", "-C", remoteDir}

	execErr := s.exec.Execute(cmd, ExecuteOpts{Stdin: reader})
	tarErr := <-tarWritingErrCh

	if tarErr != nil {
		return tarErr
	}

	return execErr
}

func (s DirCp) Down(localDir, remoteDir string) error {
	reader, writer := io.Pipe()
	tarReadingErrCh := make(chan error)

	// tar strips the leading '/' if it's there, so we will too
	tarPrefix := path.Clean(strings.TrimLeft(remoteDir, "/"))

	go func() {
		err := Tarball{reader, tarPrefix}.Extract(localDir)
		reader.Close()
		tarReadingErrCh <- err
	}()

	cmd := []string{"tar", "cf", "-", remoteDir}

	execErr := s.exec.Execute(cmd, ExecuteOpts{Stdout: writer})
	tarErr := <-tarReadingErrCh

	if tarErr != nil {
		return tarErr
	}

	return execErr
}
