package net

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type KubeListenerService struct {
	serviceName   string
	serviceType   corev1.ServiceType
	namespaceName string
	portStr       string
	coreClient    kubernetes.Interface

	redirected   bool
	created      bool
	originalSpec *corev1.ServiceSpec

	logTag string
	logger Logger
}

func NewKubeListenerService(
	svcName string, svcType corev1.ServiceType, nsName string,
	portStr string, coreClient kubernetes.Interface, logger Logger) *KubeListenerService {

	return &KubeListenerService{
		serviceName:   svcName,
		serviceType:   svcType,
		namespaceName: nsName,
		portStr:       portStr,
		coreClient:    coreClient,

		logTag: "KubeListenerService",
		logger: logger,
	}
}

func (s *KubeListenerService) Snapshot() error {
	if len(s.namespaceName) == 0 {
		return fmt.Errorf("Expected non-empty namespace name")
	}

	service, err := s.coreClient.CoreV1().Services(s.namespaceName).Get(s.serviceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil // nothing to snapshot
		}
		return err
	}

	s.originalSpec = service.Spec.DeepCopy()

	return nil
}

func (s *KubeListenerService) Redirect(targetPort string) error {
	port, err := strconv.Atoi(s.portStr)
	if err != nil {
		return fmt.Errorf("Converting port string '%s' to int: %s", s.portStr, err)
	}

	service, err := s.coreClient.CoreV1().Services(s.namespaceName).Get(s.serviceName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	if len(service.UID) == 0 {
		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      s.serviceName,
				Namespace: s.namespaceName,
			},
			Spec: corev1.ServiceSpec{
				Type:  s.serviceType,
				Ports: []corev1.ServicePort{},
			},
		}
	} else {
		service.Spec.Ports = []corev1.ServicePort{}
	}

	// TODO work with multiple ports
	service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
		Name:       "port0",
		Port:       int32(port),
		TargetPort: intstr.Parse(targetPort),
		Protocol:   corev1.ProtocolTCP,
	})

	service.Spec.Selector = map[string]string{netPodSelectorKey: netPodSelectorValue}

	if len(service.UID) == 0 {
		_, err = s.coreClient.CoreV1().Services(s.namespaceName).Create(service)
		if err != nil {
			return fmt.Errorf("Creating service: %s", err)
		}

		s.created = true
	} else {
		_, err = s.coreClient.CoreV1().Services(s.namespaceName).Update(service)
		if err != nil {
			return fmt.Errorf("Updating service: %s", err)
		}
	}

	s.redirected = true

	return nil
}

func (s *KubeListenerService) Revert() error {
	if !s.redirected {
		return nil
	}

	if s.created {
		err := s.coreClient.CoreV1().Services(s.namespaceName).Delete(s.serviceName, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Deleting service: %s", err)
		}

		return nil
	}

	service, err := s.coreClient.CoreV1().Services(s.namespaceName).Get(s.serviceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	service.Spec.Ports = s.originalSpec.Ports
	service.Spec.Selector = s.originalSpec.Selector

	_, err = s.coreClient.CoreV1().Services(s.namespaceName).Update(service)
	if err != nil {
		return fmt.Errorf("Updating service back to original: %s", err)
	}

	return nil
}
