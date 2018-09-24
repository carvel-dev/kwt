package kube

import (
	"io"
)

type DirCp struct {
	exec Exec
}

func NewDirCp(exec Exec) DirCp {
	return DirCp{exec}
}

func (s DirCp) Execute(srcDir, dstDir string) error {
	reader, writer := io.Pipe()
	tarWritingErrCh := make(chan error)

	go func() {
		err := TarBuilder{}.Build(srcDir, "/", TarBuilderOpts{ExcludedPaths: []string{".git"}}, writer)
		writer.Close()
		tarWritingErrCh <- err
	}()

	cmd := []string{"tar", "xf", "-", "-C", dstDir}

	execErr := s.exec.Execute(cmd, ExecuteOpts{Stdin: reader})
	tarErr := <-tarWritingErrCh

	if tarErr != nil {
		return tarErr
	}

	return execErr
}
