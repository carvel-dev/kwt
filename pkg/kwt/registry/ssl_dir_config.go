package registry

import (
	corev1 "k8s.io/api/core/v1"
)

type SSLDirConfig struct {
	envVarName string

	volumeName string
	certsPath  string

	dockerVolumeName string
	dockerCertsPath  string
}

func NewSSLDirConfig() SSLDirConfig {
	return SSLDirConfig{
		envVarName: "SSL_CERT_DIR", // golang's crypto/x509 (TODO SSL_CERT_FILE?)

		volumeName: "host-certs",
		certsPath:  "/etc/ssl/certs",

		dockerVolumeName: "docker-host-certs",
		dockerCertsPath:  "/etc/docker/certs.d",
	}
}

func (c SSLDirConfig) EnvVar() corev1.EnvVar {
	return corev1.EnvVar{
		Name:  c.envVarName,
		Value: c.certsPath,
	}
}

func (c SSLDirConfig) ContainsEnvVar(envs []corev1.EnvVar) bool {
	for _, env := range envs {
		if env.Name == c.envVarName {
			return true
		}
	}
	return false
}

func (c SSLDirConfig) VolumeMount(readOnly bool) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      c.volumeName,
		MountPath: c.certsPath,
		ReadOnly:  readOnly,
	}
}

func (c SSLDirConfig) Volume() corev1.Volume {
	return corev1.Volume{
		Name: c.volumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{Path: c.certsPath},
		},
	}
}

func (c SSLDirConfig) DockerVolumeMount(readOnly bool) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      c.dockerVolumeName,
		MountPath: c.dockerCertsPath,
		ReadOnly:  readOnly,
	}
}

func (c SSLDirConfig) DockerVolume() corev1.Volume {
	return corev1.Volume{
		Name: c.dockerVolumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{Path: c.dockerCertsPath},
		},
	}
}
