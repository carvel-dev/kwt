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

package kube

import (
	"bytes"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Exec struct {
	pod        corev1.Pod
	container  string
	coreClient kubernetes.Interface
	restConfig *rest.Config
}

type ExecuteOpts struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func NewExec(pod corev1.Pod, container string, coreClient kubernetes.Interface, restConfig *rest.Config) Exec {
	return Exec{pod, container, coreClient, restConfig}
}

func (s Exec) Execute(cmd []string, opts ExecuteOpts) error {
	req := s.coreClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(s.pod.Name).
		Namespace(s.pod.Namespace).
		SubResource("exec")

	// Avoid 'you must specify at least 1 of stdin, stdout, stderr' error
	var stderrBuf bytes.Buffer

	if opts.Stderr == nil {
		opts.Stderr = &stderrBuf
	}

	req.VersionedParams(&corev1.PodExecOptions{
		Stdout:    opts.Stdout != nil,
		Stderr:    opts.Stderr != nil,
		Stdin:     opts.Stdin != nil,
		TTY:       false,
		Command:   cmd,
		Container: s.container,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(s.restConfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("Building executor: %s", err)
	}

	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Stdin:  opts.Stdin,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("Execution error: %s (stderr: %s [optional])", err, stderrBuf.String())
	}

	return nil
}
