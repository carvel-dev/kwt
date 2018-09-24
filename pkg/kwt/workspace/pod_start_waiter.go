package workspace

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodStartWaiter struct {
	pod        corev1.Pod
	coreClient kubernetes.Interface
}

func (l PodStartWaiter) WaitForStart(cancelCh chan struct{}) (corev1.PodPhase, error) {
	for {
		// TODO infinite retry

		pod, err := l.coreClient.CoreV1().Pods(l.pod.Namespace).Get(l.pod.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		switch pod.Status.Phase {
		case corev1.PodRunning, corev1.PodSucceeded, corev1.PodFailed, corev1.PodUnknown:
			return pod.Status.Phase, nil
		}

		select {
		case <-cancelCh:
			return pod.Status.Phase, nil
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
