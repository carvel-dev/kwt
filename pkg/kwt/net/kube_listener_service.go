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
	coreClient    kubernetes.Interface

	redirected   bool
	newService   bool
	originalSpec *corev1.ServiceSpec

	logTag string
	logger Logger
}

func NewKubeListenerService(
	svcName string, svcType corev1.ServiceType, nsName string,
	coreClient kubernetes.Interface, logger Logger) *KubeListenerService {

	return &KubeListenerService{
		serviceName:   svcName,
		serviceType:   svcType,
		namespaceName: nsName,
		coreClient:    coreClient,

		logTag: "KubeListenerService",
		logger: logger,
	}
}

func (s *KubeListenerService) Redirect(portStr, targetPort string) error {
	if len(s.namespaceName) == 0 {
		return fmt.Errorf("Expected non-empty namespace name")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("Converting port string '%s' to int: %s", portStr, err)
	}

	service, err := s.coreClient.CoreV1().Services(s.namespaceName).Get(s.serviceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			s.newService = true
		} else {
			return err
		}
	}

	if s.newService {
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
		s.originalSpec = service.Spec.DeepCopy()
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

	if s.newService {
		_, err = s.coreClient.CoreV1().Services(s.namespaceName).Create(service)
		if err != nil {
			return fmt.Errorf("Creating service: %s", err)
		}
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

	if s.newService {
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
