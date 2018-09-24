package registry

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Info struct {
	Address  string
	Username string
	Password string
	CA       string
}

func (r Registry) Info() (Info, error) {
	r.logger.Info(r.logTag, "Waiting for service IP")

	ip, err := r.waitForServiceIP()
	if err != nil {
		return Info{}, err
	}

	authSecret, err := r.coreClient.CoreV1().Secrets(r.namespace).Get(r.authSecretName, metav1.GetOptions{})
	if err != nil {
		return Info{}, fmt.Errorf("Getting auth secret: %s", err)
	}

	tlsSecret, err := r.coreClient.CoreV1().Secrets(r.namespace).Get(r.tlsSecretName, metav1.GetOptions{})
	if err != nil {
		return Info{}, fmt.Errorf("Getting TLS secret: %s", err)
	}

	info := Info{
		Address:  "https://" + ip,
		Username: string(authSecret.Data[corev1.BasicAuthUsernameKey]),
		Password: string(authSecret.Data[corev1.BasicAuthPasswordKey]),
		CA:       string(tlsSecret.Data[tlsCACertKey]),
	}

	return info, nil
}
