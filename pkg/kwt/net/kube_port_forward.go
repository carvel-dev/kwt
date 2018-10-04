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

	outBuf *bytes.Buffer
	errBuf *bytes.Buffer
	stopCh chan struct{}

	logTag string
	logger Logger
}

func NewKubePortForward(
	pod *corev1.Pod,
	coreClient kubernetes.Interface,
	restConfig *rest.Config,
	logger Logger,
) *KubePortForward {
	return &KubePortForward{
		pod:        pod,
		coreClient: coreClient,
		restConfig: restConfig,

		outBuf: bytes.NewBufferString(""),
		errBuf: bytes.NewBufferString(""),
		stopCh: make(chan struct{}),

		logTag: "KubePortForward",
		logger: logger,
	}
}

func (f KubePortForward) Start(remotePort int, startedCh chan struct{}) error {
	req := f.coreClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(f.pod.Namespace).
		Name(f.pod.Name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(f.restConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	fw, err := portforward.New(dialer, []string{"0:" + strconv.Itoa(remotePort)}, f.stopCh, startedCh, f.outBuf, f.errBuf)
	if err != nil {
		return err
	}

	f.logger.Debug(f.logTag, "Starting port forwarding")

	err = fw.ForwardPorts()

	f.logger.Debug(f.logTag, "Finished port forwarding (err: %s)", err)

	return err
}

// LocalPort should only be called once and after port forwarding is ready
func (f KubePortForward) LocalPort() (int, error) {
	outBufResult := f.outBuf.String()
	errBufResult := f.errBuf.String()

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

	return localPort, nil
}

func (f KubePortForward) Shutdown() error {
	close(f.stopCh)
	return nil
}
