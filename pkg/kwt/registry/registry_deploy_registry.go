package registry

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apires "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const registryImage = "registry:2.6.2"

func (r Registry) deployRegistry() error {
	podSpec := corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyAlways,
		InitContainers: []corev1.Container{
			{
				Name:    "init-config",
				Image:   registryImage,
				Command: []string{"/bin/sh"},
				Args: []string{
					"-c",
					`set -e
			    echo "$REGISTRY_HTTP_TLS_CERTIFICATE" > /var/lib/registry-config/tls-cert
			    echo "$REGISTRY_HTTP_TLS_KEY" > /var/lib/registry-config/tls-key
					htpasswd -Bbn "$AUTH_USERNAME" "$AUTH_PASSWORD" > /var/lib/registry-config/htpasswd`,
				},
				Env: []corev1.EnvVar{
					{
						Name: "REGISTRY_HTTP_TLS_CERTIFICATE",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: r.tlsSecretName},
								Key:                  corev1.TLSCertKey,
							},
						},
					},
					{
						Name: "REGISTRY_HTTP_TLS_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: r.tlsSecretName},
								Key:                  corev1.TLSPrivateKeyKey,
							},
						},
					},
					{
						Name: "AUTH_USERNAME",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: r.authSecretName},
								Key:                  corev1.BasicAuthUsernameKey,
							},
						},
					},
					{
						Name: "AUTH_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: r.authSecretName},
								Key:                  corev1.BasicAuthPasswordKey,
							},
						},
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "config-store", MountPath: "/var/lib/registry-config"},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "registry",
				Image: registryImage,
				Resources: corev1.ResourceRequirements{
					// keep request==limit to keep this container in guaranteed class
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    apires.MustParse("100m"),
						corev1.ResourceMemory: apires.MustParse("100Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    apires.MustParse("100m"),
						corev1.ResourceMemory: apires.MustParse("100Mi"),
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "REGISTRY_HTTP_ADDR",
						Value: ":443",
					},
					{
						Name:  "REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY",
						Value: "/var/lib/registry-images",
					},
					{
						Name:  "REGISTRY_HTTP_TLS_CERTIFICATE",
						Value: "/var/lib/registry-config/tls-cert",
					},
					{
						Name:  "REGISTRY_HTTP_TLS_KEY",
						Value: "/var/lib/registry-config/tls-key",
					},
					{
						Name:  "REGISTRY_AUTH",
						Value: "htpasswd",
					},
					{
						Name:  "REGISTRY_AUTH_HTPASSWD_REALM",
						Value: "Registry Realm",
					},
					{
						Name:  "REGISTRY_AUTH_HTPASSWD_PATH",
						Value: "/var/lib/registry-config/htpasswd",
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "config-store", MountPath: "/var/lib/registry-config"},
					{Name: "image-store", MountPath: "/var/lib/registry-images"},
				},
				Ports: []corev1.ContainerPort{
					{ContainerPort: 443, Name: "registry", Protocol: corev1.ProtocolTCP},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name:         "config-store",
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
			{
				Name:         "image-store",
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
		},
	}

	var replicas int32 = 1

	rc := &corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.registryRCName,
			Namespace: r.namespace,
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				serviceSelectorKey: serviceSelectorValue,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						serviceSelectorKey: serviceSelectorValue,
					},
				},
				Spec: podSpec,
			},
		},
	}

	_, err := r.coreClient.CoreV1().ReplicationControllers(r.namespace).Create(rc)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Creating RC: %s", err)
		}
	}

	return nil
}

func (r Registry) waitForRegistry() error {
	timeoutCh := time.After(2 * time.Minute)

	for {
		rc, err := r.coreClient.CoreV1().ReplicationControllers(r.namespace).Get(r.registryRCName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if rc.Status.Replicas == rc.Status.ReadyReplicas {
			return nil
		}

		select {
		case <-timeoutCh:
			return fmt.Errorf("Timed out waiting for Registry RC '%s' to have all pods running", r.registryRCName)
		default:
			// continue with waiting
		}

		time.Sleep(1 * time.Second)
	}
}
