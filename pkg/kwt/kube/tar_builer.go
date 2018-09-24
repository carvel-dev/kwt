/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Adapted from https://github.com/kubernetes/kubernetes/blob/9f8ebd171479bec0ada837d7ee641dec2f8c6dd1/pkg/kubectl/cmd/cp.go

package kube

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type TarBuilder struct{}

type TarBuilderOpts struct {
	ExcludedPaths []string

	ReportFileExcluded func(string)
	ReportFileIncluded func(string)
}

type tarBuilderInternal struct {
	Writer        *tar.Writer
	ExcludedPaths map[string]struct{}

	ReportFileExcluded func(string)
	ReportFileIncluded func(string)
}

func (b TarBuilder) Build(srcPath, destPath string, opts TarBuilderOpts, writer io.Writer) error {
	tarWriter := tar.NewWriter(writer) // TODO compress?
	defer tarWriter.Close()

	srcPath = path.Clean(srcPath)
	destPath = path.Clean(destPath)

	internal := tarBuilderInternal{
		Writer:        tarWriter,
		ExcludedPaths: map[string]struct{}{},

		ReportFileExcluded: opts.ReportFileExcluded,
		ReportFileIncluded: opts.ReportFileIncluded,
	}

	if internal.ReportFileExcluded == nil {
		internal.ReportFileExcluded = func(string) {}
	}
	if internal.ReportFileIncluded == nil {
		internal.ReportFileIncluded = func(string) {}
	}

	for _, ep := range opts.ExcludedPaths {
		excludedPath := path.Clean(path.Join(srcPath, ep))
		internal.ReportFileExcluded(excludedPath)
		internal.ExcludedPaths[excludedPath] = struct{}{}
	}

	return b.recursiveTar(path.Dir(srcPath), path.Base(srcPath), path.Dir(destPath), path.Base(destPath), internal)
}

func (b TarBuilder) recursiveTar(srcBase, srcFile, destBase, destFile string, internal tarBuilderInternal) error {
	filepath := path.Join(srcBase, srcFile)
	stat, err := os.Lstat(filepath)
	if err != nil {
		return err
	}
	if _, found := internal.ExcludedPaths[filepath]; found {
		return nil
	}
	internal.ReportFileIncluded(filepath)
	if stat.IsDir() {
		files, err := ioutil.ReadDir(filepath)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			//case empty directory
			hdr, _ := tar.FileInfoHeader(stat, filepath)
			hdr.Name = destFile
			if err := internal.Writer.WriteHeader(hdr); err != nil {
				return err
			}
		}
		for _, f := range files {
			if err := b.recursiveTar(srcBase, path.Join(srcFile, f.Name()), destBase, path.Join(destFile, f.Name()), internal); err != nil {
				return err
			}
		}
		return nil
	} else if stat.Mode()&os.ModeSymlink != 0 {
		//case soft link
		hdr, _ := tar.FileInfoHeader(stat, filepath)
		target, err := os.Readlink(filepath)
		if err != nil {
			return err
		}

		hdr.Linkname = target
		hdr.Name = destFile
		if err := internal.Writer.WriteHeader(hdr); err != nil {
			return err
		}
	} else {
		//case regular file or other file type like pipe
		hdr, err := tar.FileInfoHeader(stat, filepath)
		if err != nil {
			return err
		}
		hdr.Name = destFile

		if err := internal.Writer.WriteHeader(hdr); err != nil {
			return err
		}

		f, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(internal.Writer, f); err != nil {
			return err
		}
		return f.Close()
	}
	return nil
}
