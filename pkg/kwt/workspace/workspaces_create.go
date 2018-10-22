package workspace

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	workspaceContainerName = "debug"
)

type CreateOpts struct {
	Name         string
	GenerateName bool

	Image       string
	Command     []string
	CommandArgs []string
	Privileged  bool

	Ports []int

	ServiceAccountName string
}

func (w Workspaces) Create(opts CreateOpts) (Workspace, error) {
	trueBool := true
	zeroInt64 := int64(0)

	pod := &corev1.Pod{
		ObjectMeta: w.applyGenerateName(opts.GenerateName, metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: w.namespace,
			Labels: map[string]string{
				workspaceLabel: "true",
			},
		}),
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,

			// Requires alpha feature in Kubernetes 1.11
			// https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/
			// (once enabled can access filesystem as well: /proc/$pid/root)
			// TODO ShareProcessNamespace: &trueBool,

			Containers:         []corev1.Container{{}},
			ServiceAccountName: opts.ServiceAccountName,
		},
	}

	if len(opts.Image) > 0 {
		pod.Spec.Containers[0] = corev1.Container{
			Name: workspaceContainerName,

			Image:           opts.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,

			Command: opts.Command,
			Args:    opts.CommandArgs,
		}
	} else {
		pod.Spec.Containers[0] = corev1.Container{
			Name: workspaceContainerName,

			Image:           "ubuntu:xenial",
			ImagePullPolicy: corev1.PullIfNotPresent,

			Command: []string{"/bin/bash"},
			Args:    []string{"-c", "while true; do sleep 86400; done"}, // sleep forever

			WorkingDir: ContainerEnv{}.WorkingDir(),
		}
	}

	if len(opts.Ports) > 0 {
		for i, port := range opts.Ports {
			pod.Spec.Containers[0].Ports = append(pod.Spec.Containers[0].Ports,
				corev1.ContainerPort{Name: "port" + strconv.Itoa(i), ContainerPort: int32(port)})
		}
	}

	if opts.Privileged {
		for i, _ := range pod.Spec.Containers {
			pod.Spec.Containers[i].SecurityContext = &corev1.SecurityContext{
				Privileged: &trueBool,
				RunAsUser:  &zeroInt64, // 0 is root
			}
		}
	}

	createdPod, err := w.coreClient.CoreV1().Pods(w.namespace).Create(pod)
	if err != nil {
		return nil, err
	}

	workspace := &WorkspaceImpl{*createdPod, w.coreClient}

	return workspace, nil
}

func (w Workspaces) applyGenerateName(generate bool, meta metav1.ObjectMeta) metav1.ObjectMeta {
	if generate {
		meta.GenerateName = meta.Name + "-"
		meta.Name = ""
	}
	return meta
}
