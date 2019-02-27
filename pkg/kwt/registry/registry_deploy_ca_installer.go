package registry

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apires "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deployCAInstaller creates a DaemonSet which installs registry's CA certificate
// on hosts so that container runtime (Docker) can use cluster registry for pulling images.
func (r Registry) deployCAInstaller(ip string) error {
	sslConfig := NewSSLDirConfig()
	// trueBool := true
	// zeroInt64 := int64(0)

	podSpec := corev1.PodSpec{
		// HostPID:       true,
		RestartPolicy: corev1.RestartPolicyAlways,
		Containers: []corev1.Container{
			{
				Name:    "install-certs",
				Image:   registryImage,
				Command: []string{"/bin/sh"},
				Args: []string{
					"-c",
					fmt.Sprintf(`set -e

          echo "$REGISTRY_HTTP_TLS_CA_CERTIFICATE" > /etc/ssl/certs/kwt-registry-ca.pem

          ip="%s"
          mkdir -p /etc/docker/certs.d/${ip}
          echo "$REGISTRY_HTTP_TLS_CA_CERTIFICATE" > /etc/docker/certs.d/${ip}/ca.pem

          mkdir -p /etc/docker/certs.d/${ip}:443
          echo "$REGISTRY_HTTP_TLS_CA_CERTIFICATE" > /etc/docker/certs.d/${ip}:443/ca.pem

          while true; do sleep 86000; done`, ip),
				},
				Resources: corev1.ResourceRequirements{
					// keep request==limit to keep this container in guaranteed class
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    apires.MustParse("100m"),
						corev1.ResourceMemory: apires.MustParse("50Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    apires.MustParse("100m"),
						corev1.ResourceMemory: apires.MustParse("50Mi"),
					},
				},
				Env: []corev1.EnvVar{
					{
						Name: "REGISTRY_HTTP_TLS_CA_CERTIFICATE",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: r.tlsSecretName},
								Key:                  corev1.TLSCertKey,
							},
						},
					},
				},
				// SecurityContext: &corev1.SecurityContext{
				// 	Privileged: &trueBool,
				// 	RunAsUser:  &zeroInt64, // 0 is root
				// },
				VolumeMounts: []corev1.VolumeMount{
					sslConfig.VolumeMount(false),
					sslConfig.DockerVolumeMount(false),
				},
			},
		},
		Volumes: []corev1.Volume{
			sslConfig.Volume(),
			sslConfig.DockerVolume(),
		},
	}

	const (
		installerSelectorKey   = "kwt-registry-ca-installer" // TODO namespace
		installerSelectorValue = "1"
	)

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.installerDSName,
			Namespace: r.namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					installerSelectorKey: installerSelectorValue,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						installerSelectorKey: installerSelectorValue,
					},
				},
				Spec: podSpec,
			},
		},
	}

	_, err := r.coreClient.AppsV1().DaemonSets(r.namespace).Create(ds)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Creating DS: %s", err)
		}
	}

	return nil
}

func (r Registry) waitForCAInstaller() error {
	timeoutCh := time.After(2 * time.Minute)

	for {
		ds, err := r.coreClient.AppsV1().DaemonSets(r.namespace).Get(r.installerDSName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if ds.Status.DesiredNumberScheduled == ds.Status.NumberReady {
			return nil
		}

		select {
		case <-timeoutCh:
			return fmt.Errorf("Timed out waiting for CA installer DS '%s' to have all pods running", r.installerDSName)
		default:
			// continue with waiting
		}

		time.Sleep(1 * time.Second)
	}
}
