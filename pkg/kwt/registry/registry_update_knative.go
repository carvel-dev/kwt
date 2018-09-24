package registry

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const knativeServingNs = "knative-serving"

type KnativeServing struct {
	coreClient kubernetes.Interface
}

func (r KnativeServing) IsInstalled() (bool, error) {
	_, err := r.coreClient.CoreV1().Namespaces().Get(knativeServingNs, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("Getting namespace: %s", err)
	}
	return true, nil
}

// UpdateControllerToUseHostCAs updates Knative Serving controller to pick up host CA certificates
// (which includes registry CA certificate) so that controller can resolve image digests
// by communicating with the registry. Eventually this would not be necessary as Knative Build
// will provide image digests as part of build result.
func (r KnativeServing) UpdateControllerToUseHostCAs() error {
	foundDep, err := r.coreClient.ExtensionsV1beta1().Deployments(knativeServingNs).Get("controller", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Getting deployment: %s", err)
	}

	sslConfig := NewSSLDirConfig()

	if sslConfig.ContainsEnvVar(foundDep.Spec.Template.Spec.Containers[0].Env) {
		// If deployment is already configured to use host certificates,
		// force restart of deployment managed pods so that controller
		// loads potentially new set of certificates during its startup.
		// (https://github.com/kubernetes/kubernetes/issues/27081 and https://github.com/kubernetes/kubernetes/issues/13488)
		foundDep.Spec.Template.Annotations["kwt.cppforlife.io/force-update"] = time.Now().UTC().Format(time.RFC3339Nano)

		_, err = r.coreClient.ExtensionsV1beta1().Deployments(knativeServingNs).Update(foundDep)
		if err != nil {
			return fmt.Errorf("Updating deployment metadata to force restart: %s", err)
		}

		return nil
	}

	for i, cont := range foundDep.Spec.Template.Spec.Containers {
		cont.Env = append(cont.Env, sslConfig.EnvVar())
		cont.VolumeMounts = append(cont.VolumeMounts, sslConfig.VolumeMount(true))
		foundDep.Spec.Template.Spec.Containers[i] = cont
	}

	foundDep.Spec.Template.Spec.Volumes = append(foundDep.Spec.Template.Spec.Volumes, sslConfig.Volume())

	_, err = r.coreClient.ExtensionsV1beta1().Deployments(knativeServingNs).Update(foundDep)
	if err != nil {
		return fmt.Errorf("Updating deployment: %s", err)
	}

	return nil
}
