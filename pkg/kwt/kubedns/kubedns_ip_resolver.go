package kubedns

import (
	"fmt"
	"net"
	"strings"

	ctldns "github.com/cppforlife/kwt/pkg/kwt/dns"
	ctlmdns "github.com/cppforlife/kwt/pkg/kwt/mdns"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	clusterSuffix = ".cluster.local."
	svcSuffix     = ".svc" + clusterSuffix
	podSuffix     = ".pod" + clusterSuffix
)

type KubeDNSIPResolver struct {
	coreClient kubernetes.Interface
}

var _ ctldns.IPResolver = KubeDNSIPResolver{}
var _ ctlmdns.IPResolver = KubeDNSIPResolver{}

func NewKubeDNSIPResolver(coreClient kubernetes.Interface) KubeDNSIPResolver {
	return KubeDNSIPResolver{coreClient}
}

func (r KubeDNSIPResolver) String() string { return "kube-dns" }

func (r KubeDNSIPResolver) ResolveIPv4(question string) ([]net.IP, bool, error) {
	if !strings.HasSuffix(question, clusterSuffix) {
		return nil, false, nil
	}

	// TODO following implementation is incomplete
	// see more: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
	switch {
	// smy-svc.my-namespace.svc.cluster.local -> cluster IP
	case strings.HasSuffix(question, svcSuffix):
		ips, err := r.svcIP(question)
		return ips, true, err

	// 1-2-3-4.default.pod.cluster.local -> 1.2.3.4
	case strings.HasSuffix(question, podSuffix):
		ips, err := r.podIP(question)
		return ips, true, err

	default:
		return nil, true, fmt.Errorf("Could not determine Kubernetes DNS question format")
	}
}

func (r KubeDNSIPResolver) ResolveIPv6(question string) ([]net.IP, bool, error) {
	if !strings.HasSuffix(question, clusterSuffix) {
		return nil, false, nil
	}

	switch {
	case strings.HasSuffix(question, svcSuffix):
		return nil, true, nil // no IPs

	case strings.HasSuffix(question, podSuffix):
		return nil, true, nil // no IPs

	default:
		return nil, true, fmt.Errorf("Could not determine Kubernetes DNS question format")
	}
}

func (r KubeDNSIPResolver) svcIP(question string) ([]net.IP, error) {
	rest := strings.TrimSuffix(question, svcSuffix)

	pieces := strings.SplitN(rest, ".", 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("Expected service address to be in particular format")
	}

	svc, err := r.coreClient.CoreV1().Services(pieces[1]).Get(pieces[0], metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Getting service: %s", err)
	}

	if len(svc.Spec.ClusterIP) == 0 {
		// TODO which pods to pick up (readiness? etc?)
		return nil, nil
	}

	ip := net.ParseIP(svc.Spec.ClusterIP)
	if ip == nil {
		return nil, fmt.Errorf("Expected service cluster IP address to be valid")
	}

	return []net.IP{ip}, nil
}

func (r KubeDNSIPResolver) podIP(question string) ([]net.IP, error) {
	rest := strings.TrimSuffix(question, podSuffix)

	pieces := strings.SplitN(rest, ".", 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("Expected pod address to be in particular format")
	}

	ip := net.ParseIP(strings.Replace(pieces[0], "-", ".", -1))
	if ip == nil {
		return nil, fmt.Errorf("Expected pod address to be in IP address format")
	}

	return []net.IP{ip}, nil
}

func ServiceInternalDNSAddress(service corev1.Service) string {
	addr := fmt.Sprintf("%s.%s%s", service.Name, service.Namespace, svcSuffix)
	return strings.TrimSuffix(addr, ".")
}

func PodInternalDNSAddress(pod corev1.Pod) string {
	if len(pod.Status.PodIP) > 0 {
		dashedIP := strings.Replace(pod.Status.PodIP, ".", "-", -1)
		addr := fmt.Sprintf("%s.%s%s", dashedIP, pod.Namespace, podSuffix)
		return strings.TrimSuffix(addr, ".")
	}
	return ""
}
