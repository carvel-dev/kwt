package net

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	sshSecretPublicKey = "ssh-publickey"
)

type KubeEntryPoint struct {
	coreClient kubernetes.Interface
	restConfig *rest.Config

	namespace string
	podName   string
	podPort   int

	secretClientSSHName string
	secretHostSSHName   string

	logTag string
	logger Logger
}

var _ EntryPoint = KubeEntryPoint{}

func NewKubeEntryPoint(coreClient kubernetes.Interface, restConfig *rest.Config, namespace string, logger Logger) KubeEntryPoint {
	return KubeEntryPoint{
		coreClient: coreClient,
		restConfig: restConfig,

		namespace: namespace,
		podName:   "kwt-net",
		podPort:   2048,

		secretClientSSHName: "kwt-net-ssh-key",
		secretHostSSHName:   "kwt-net-host-key",

		logTag: "KubeEntryPoint",
		logger: logger,
	}
}

func (f KubeEntryPoint) EntryPoint() (dstconn.SSHClientConnOpts, error) {
	opts := dstconn.SSHClientConnOpts{}

	var clientPrivateKeyPEM, hostPublicKeyAuf string
	sshKeysErrCh := make(chan error)

	f.logger.Info(f.logTag, "Creating networking client secret '%s' in namespace '%s'...", f.secretClientSSHName, f.namespace)

	go func() {
		var err error
		clientPrivateKeyPEM, err = f.createNetPodClientSSHSecret()
		sshKeysErrCh <- err
	}()

	f.logger.Info(f.logTag, "Creating networking host secret '%s' in namespace '%s'...", f.secretHostSSHName, f.namespace)

	go func() {
		var err error
		hostPublicKeyAuf, err = f.createNetPodHostSSHSecret()
		sshKeysErrCh <- err
	}()

	for i := 0; i < 2; i++ {
		err := <-sshKeysErrCh
		if err != nil {
			return opts, err
		}
	}

	f.logger.Info(f.logTag, "Creating networking pod '%s' in namespace '%s'", f.podName, f.namespace)

	pod, err := f.createNetPod()
	if err != nil {
		return opts, err
	}

	f.logger.Info(f.logTag, "Waiting for networking pod '%s' in namespace '%s' to start...", f.podName, f.namespace)

	ready, err := f.waitForPod(pod)
	if err != nil {
		return opts, err
	}

	if ready {
		pf := NewKubePortForward(pod, f.coreClient, f.restConfig, f.logger)

		localPort, err := pf.Start(f.podPort)
		if err != nil {
			return opts, fmt.Errorf("Starting kube port forwarding: %s", err)
		}

		opts := dstconn.SSHClientConnOpts{
			User:             "tom",
			Host:             "localhost:" + strconv.Itoa(localPort),
			PrivateKeyPEM:    clientPrivateKeyPEM,
			HostPublicKeyAuf: hostPublicKeyAuf,
		}

		return opts, nil
	}

	return opts, fmt.Errorf("Network pod failed to start")
}

func (f KubeEntryPoint) Delete() error {
	err := f.coreClient.CoreV1().Pods(f.namespace).Delete(f.podName, &metav1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("Deleting net pod: %s", err)
		}
	}

	err = f.coreClient.CoreV1().Secrets(f.namespace).Delete(f.secretClientSSHName, &metav1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("Deleting net pod client ssh secret: %s", err)
		}
	}

	err = f.coreClient.CoreV1().Secrets(f.namespace).Delete(f.secretHostSSHName, &metav1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("Deleting net pod host ssh secret: %s", err)
		}
	}

	err = f.waitForObjDeletion(fmt.Sprintf("pod '%s'", f.podName), func() error {
		_, err := f.coreClient.CoreV1().Pods(f.namespace).Get(f.podName, metav1.GetOptions{})
		return err
	})
	if err != nil {
		return err
	}

	err = f.waitForObjDeletion(fmt.Sprintf("secret '%s'", f.secretClientSSHName), func() error {
		_, err := f.coreClient.CoreV1().Secrets(f.namespace).Get(f.secretClientSSHName, metav1.GetOptions{})
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (f KubeEntryPoint) createNetPodClientSSHSecret() (string, error) {
	foundSecret, err := f.coreClient.CoreV1().Secrets(f.namespace).Get(f.secretClientSSHName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return "", fmt.Errorf("Getting net pod client ssh secret: %s", err)
		}
		// continue to create new ssh key
	} else {
		return string(foundSecret.Data[corev1.SSHAuthPrivateKey]), nil
	}

	key, err := dstconn.NewSSHKeyGenerator().Generate()
	if err != nil {
		return "", err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.secretClientSSHName,
			Namespace: f.namespace,
		},
		Type: corev1.SecretTypeSSHAuth,
		StringData: map[string]string{
			corev1.SSHAuthPrivateKey: key.PrivateKey,
			sshSecretPublicKey:       key.PublicKey,
		},
	}

	_, err = f.coreClient.CoreV1().Secrets(f.namespace).Create(secret)
	if err != nil {
		return "", fmt.Errorf("Creating net pod client ssh secret: %s", err)
	}

	return key.PrivateKey, nil
}

func (f KubeEntryPoint) createNetPodHostSSHSecret() (string, error) {
	foundSecret, err := f.coreClient.CoreV1().Secrets(f.namespace).Get(f.secretHostSSHName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return "", fmt.Errorf("Getting net pod host ssh secret: %s", err)
		}
		// continue to create new ssh key
	} else {
		return string(foundSecret.Data[sshSecretPublicKey]), nil
	}

	key, err := dstconn.NewSSHKeyGenerator().Generate()
	if err != nil {
		return "", err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.secretHostSSHName,
			Namespace: f.namespace,
		},
		Type: corev1.SecretTypeSSHAuth,
		StringData: map[string]string{
			corev1.SSHAuthPrivateKey: key.PrivateKey,
			sshSecretPublicKey:       key.PublicKey,
		},
	}

	_, err = f.coreClient.CoreV1().Secrets(f.namespace).Create(secret)
	if err != nil {
		return "", fmt.Errorf("Creating net pod host ssh secret: %s", err)
	}

	return key.PublicKey, nil
}

func (f KubeEntryPoint) createNetPod() (*corev1.Pod, error) {
	foundPod, err := f.coreClient.CoreV1().Pods(f.namespace).Get(f.podName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("Getting net pod: %s", err)
		}
		// continue to create new pod
	} else {
		if foundPod.DeletionTimestamp != nil {
			return nil, fmt.Errorf("Networking pod is currently being terminated")
		}
		return foundPod, nil
	}

	container := corev1.Container{
		Name: f.podName,

		Image:           "registry.hub.docker.com/cppforlife/sshd@sha256:f9427e82765e3fc0a7ef1357f00e64cb8754dba8370b2a6176431b8b6f48b85b",
		ImagePullPolicy: corev1.PullAlways,

		// Locally, `cd images/sshd && docker build . -t cppforlife/sshd:latest`
		// Image:           "cppforlife/sshd:latest",
		// ImagePullPolicy: corev1.PullNever,

		Command: []string{"/bin/bash"},
		Args: []string{
			"-c",
			strings.Join([]string{
				// TODO move to init container
				`echo "$KWT_CLIENT_PUB_KEY" > ~/.ssh/authorized_keys`,
				`echo "$KWT_HOST_PRIV_KEY" > /etc/ssh/ssh_host_rsa_key`,
				`echo "$KWT_HOST_PUB_KEY" > /etc/ssh/ssh_host_rsa_key.pub`,
				fmt.Sprintf("exec /usr/sbin/sshd -D -p %d", f.podPort),
			}, " && "),
		},

		Ports: []corev1.ContainerPort{
			{Name: "ssh", ContainerPort: int32(f.podPort)},
		},

		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(f.podPort),
				},
			},
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
		},

		Env: []corev1.EnvVar{
			{
				Name: "KWT_CLIENT_PUB_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: f.secretClientSSHName,
						},
						Key: sshSecretPublicKey,
					},
				},
			},
			{
				Name: "KWT_HOST_PRIV_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "kwt-net-host-key",
						},
						Key: corev1.SSHAuthPrivateKey,
					},
				},
			},
			{
				Name: "KWT_HOST_PUB_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "kwt-net-host-key",
						},
						Key: sshSecretPublicKey,
					},
				},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.podName,
			Namespace: f.namespace,
			Annotations: map[string]string{
				"sidecar.istio.io/inject": "false", // just in case Istio is used
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			Containers:    []corev1.Container{container},
		},
	}

	createdPod, err := f.coreClient.CoreV1().Pods(f.namespace).Create(pod)
	if err != nil {
		return nil, fmt.Errorf("Creating net pod: %s", err)
	}

	return createdPod, nil
}

func (f KubeEntryPoint) waitForPod(pod *corev1.Pod) (bool, error) {
	timeoutCh := time.After(2 * time.Minute)
	notifiedOfRestarts := false

	for {
		pod, err := f.coreClient.CoreV1().Pods(f.namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		podRunning := pod.Status.Phase == corev1.PodRunning
		podReady := f.podReadyConditionStatus(pod) == corev1.ConditionTrue // readiness probe

		if podRunning && podReady {
			return true, nil
		}

		if !notifiedOfRestarts {
			for _, contStatus := range pod.Status.ContainerStatuses {
				if contStatus.Name == pod.Name {
					if contStatus.RestartCount > 0 {
						notifiedOfRestarts = true
						f.logger.Error(f.logTag, "Networking pod '%s' in namespace '%s' is restarting which "+
							"may mean it encountered a problem! Continuing to wait...", pod.Name, f.namespace)
					}
					break
				}
			}
		}

		select {
		case <-timeoutCh:
			return false, fmt.Errorf("Timed out waiting for networking pod to be running/ready")
		default:
			// continue with waiting
		}

		time.Sleep(1 * time.Second)
	}
}

func (f KubeEntryPoint) podReadyConditionStatus(pod *corev1.Pod) corev1.ConditionStatus {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status
		}
	}
	return corev1.ConditionUnknown
}

func (f KubeEntryPoint) waitForObjDeletion(objDesc string, tryFunc func() error) error {
	timeoutCh := time.After(2 * time.Minute)

	for {
		err := tryFunc()
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
		}

		select {
		case <-timeoutCh:
			return fmt.Errorf("Timed out waiting for %s to be deleted", objDesc)
		default:
			// continue with waiting
		}

		time.Sleep(1 * time.Second)
	}
}
