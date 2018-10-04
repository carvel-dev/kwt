package net

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var localEphPortRegexp = regexp.MustCompile("Forwarding from 127.0.0.1:(\\d+)\\s")

type KubePortForward struct {
	pod        *corev1.Pod
	coreClient kubernetes.Interface
	restConfig *rest.Config

	logTag string
	logger Logger
}

func NewKubePortForward(
	pod *corev1.Pod,
	coreClient kubernetes.Interface,
	restConfig *rest.Config,
	logger Logger,
) *KubePortForward {
	return &KubePortForward{pod, coreClient, restConfig, "KubePortForward", logger}
}

func (f KubePortForward) Start(remotePort int) (int, error) {
	req := f.coreClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(f.pod.Namespace).
		Name(f.pod.Name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(f.restConfig)
	if err != nil {
		return 0, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	stopCh := make(<-chan struct{})
	readyCh := make(chan struct{})
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	fw, err := portforward.New(dialer, []string{"0:" + strconv.Itoa(remotePort)}, stopCh, readyCh, &outBuf, &errBuf)
	if err != nil {
		return 0, err
	}

	go func() {
		f.logger.Debug(f.logTag, "Starting port forwarding")
		err := fw.ForwardPorts()
		f.logger.Debug(f.logTag, "Finished port forwarding (err: %s)", err)
	}()

	<-readyCh

	outBufResult := outBuf.String()
	errBufResult := errBuf.String()

	matches := localEphPortRegexp.FindStringSubmatch(outBufResult)
	if len(matches) != 2 {
		return 0, fmt.Errorf("Failed to find local port: out: %s, err: %s", outBufResult, errBufResult)
	}

	localPort, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("Parsing local port: %s", err)
	}

	f.logger.Debug(f.logTag, "out: %s", outBufResult)
	f.logger.Debug(f.logTag, "err: %s", errBufResult)

	// TODO clean up stop

	return localPort, nil
}
