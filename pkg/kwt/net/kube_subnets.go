package net

import (
	"net"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeSubnets struct {
	coreClient          kubernetes.Interface
	additionalRemoteIPs []string

	logTag string
	logger Logger
}

func NewKubeSubnets(coreClient kubernetes.Interface, additionalRemoteIPs []string, logger Logger) KubeSubnets {
	return KubeSubnets{coreClient, additionalRemoteIPs, "KubeSubnets", logger}
}

func (s KubeSubnets) Subnets() ([]net.IPNet, error) {
	localIPs, err := LocalIPs()
	if err != nil {
		return nil, err
	}

	t1 := time.Now()

	podIPs, err := s.podIPs()
	if err != nil {
		return nil, err
	}

	svcIPs, err := s.serviceIPs()
	if err != nil {
		return nil, err
	}

	t2 := time.Now()

	s.logger.Debug(s.logTag, "Finished fetching pods (%d) and services (%d) in %s", len(podIPs), len(svcIPs), t2.Sub(t1))

	var remoteIPs []net.IP

	remoteIPs = append(remoteIPs, podIPs...)
	remoteIPs = append(remoteIPs, svcIPs...)

	for _, ipStr := range s.additionalRemoteIPs {
		ip := net.ParseIP(ipStr)
		if ip != nil {
			remoteIPs = append(remoteIPs, ip)
		}
	}

	return GuessSubnets(remoteIPs, localIPs), nil
}

func (s KubeSubnets) podIPs() ([]net.IP, error) {
	podList, err := s.coreClient.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []net.IP

	for _, pod := range podList.Items {
		if len(pod.Status.PodIP) > 0 {
			ip := net.ParseIP(pod.Status.PodIP)
			if ip != nil {
				result = append(result, ip)
			}
		}
	}

	return result, nil
}

func (s KubeSubnets) serviceIPs() ([]net.IP, error) {
	svcList, err := s.coreClient.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []net.IP

	for _, svc := range svcList.Items {
		if len(svc.Spec.ClusterIP) > 0 {
			ip := net.ParseIP(svc.Spec.ClusterIP) // ClusterIP can be "None"
			if ip != nil {
				result = append(result, ip)
			}
		}
	}

	return result, nil
}
