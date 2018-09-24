package registry

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	tlsCACertKey    = "ca.crt"
	authHtpasswdKey = "htpasswd"

	serviceSelectorKey   = "kwt-registry" // TODO namespace?
	serviceSelectorValue = "kwt-registry"
)

type Registry struct {
	coreClient kubernetes.Interface
	namespace  string

	serviceName          string
	tlsSecretName        string
	authSecretName       string
	dockerjsonSecretName string
	registryRCName       string
	installerDSName      string

	logTag string
	logger Logger
}

func NewRegistry(coreClient kubernetes.Interface, namespace string, logger Logger) Registry {
	const prefix = "registry"

	return Registry{
		coreClient: coreClient,
		namespace:  namespace,

		serviceName:          prefix + "-ingress",
		tlsSecretName:        prefix + "-tls",
		authSecretName:       prefix + "-auth",
		dockerjsonSecretName: prefix + "-dockerjson",
		registryRCName:       prefix + "-controller",
		installerDSName:      prefix + "-ca-installer",

		logTag: "Registry",
		logger: logger,
	}
}

func (r Registry) Install() error {
	r.logger.Info(r.logTag, "Creating namespace '%s'", r.namespace)

	err := r.getOrCreateNs()
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Creating service '%s'", r.serviceName)

	err = r.getOrCreateService()
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Waiting for service IP")

	ip, err := r.waitForServiceIP()
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Creating TLS certificates secret '%s'", r.tlsSecretName)

	err = r.getOrCreateTLSCertsSecret(ip)
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Creating auth secret '%s'", r.authSecretName)

	username, password, err := r.getOrCreateAuthSecret()
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Creating image pull secret '%s'", r.dockerjsonSecretName)

	err = r.getOrCreateImagePullSecret(ip, username, password)
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Deploying Docker registry '%s'", r.registryRCName)

	err = r.deployRegistry()
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Deploying CA installer '%s'", r.installerDSName)

	err = r.deployCAInstaller()
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Waiting for Docker registry")

	err = r.waitForRegistry()
	if err != nil {
		return err
	}

	r.logger.Info(r.logTag, "Waiting for CA installer")

	// Wait installer to add CA certificates to all hosts,
	// before trying to update Knative Serving since controller doesn't reload certs after the fact
	err = r.waitForCAInstaller()
	if err != nil {
		return err
	}

	knativeServing := KnativeServing{r.coreClient}

	installed, err := knativeServing.IsInstalled()
	if err != nil {
		return err
	}

	if installed {
		r.logger.Info(r.logTag, "Updating Knative Serving controller")

		err = knativeServing.UpdateControllerToUseHostCAs()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r Registry) Uninstall() error {
	r.logger.Info(r.logTag, "Deleting namespace")

	err := r.coreClient.CoreV1().Namespaces().Delete(r.namespace, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Deleting namespace: %s", err)
	}

	return r.waitForNsDeletion()
}

func (r Registry) getOrCreateNs() error {
	_, err := r.coreClient.CoreV1().Namespaces().Get(r.namespace, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("Getting namespace: %s", err)
		}
		return r.createNs()
	}
	return nil
}

func (r Registry) createNs() error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.namespace,
		},
	}

	_, err := r.coreClient.CoreV1().Namespaces().Create(ns)
	if err != nil {
		return fmt.Errorf("Creating namespace: %s", err)
	}

	return nil
}

func (r Registry) getOrCreateService() error {
	_, err := r.coreClient.CoreV1().Services(r.namespace).Get(r.serviceName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("Getting service: %s", err)
		}
		return r.createService()
	}
	return nil
}

func (r Registry) createService() error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.serviceName,
			Namespace: r.namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{Name: "tls", Port: 443, Protocol: corev1.ProtocolTCP},
			},
			Selector: map[string]string{
				serviceSelectorKey: serviceSelectorValue,
			},
		},
	}

	_, err := r.coreClient.CoreV1().Services(r.namespace).Create(svc)
	if err != nil {
		return fmt.Errorf("Creating service: %s", err)
	}

	return nil
}

func (r Registry) waitForServiceIP() (string, error) {
	timeoutCh := time.After(2 * time.Minute)

	for {
		svc, err := r.coreClient.CoreV1().Services(r.namespace).Get(r.serviceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		if len(svc.Spec.ClusterIP) > 0 {
			ip := net.ParseIP(svc.Spec.ClusterIP)
			if ip != nil {
				return svc.Spec.ClusterIP, nil
			}
			return "", fmt.Errorf("Expected service '%s' to have valid cluaterIP", r.serviceName)
		}

		select {
		case <-timeoutCh:
			return "", fmt.Errorf("Timed out waiting for service '%s' clusterIP", r.serviceName)
		default:
			// continue with waiting
		}

		time.Sleep(1 * time.Second)
	}
}

func (r Registry) getOrCreateTLSCertsSecret(ip string) error {
	_, err := r.coreClient.CoreV1().Secrets(r.namespace).Get(r.tlsSecretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("Getting TLS secret: %s", err)
		}
		return r.createTLSCertsSecret(ip)
	}
	return nil
}

func (r Registry) createTLSCertsSecret(ip string) error {
	store := NewInMemoryCertificateStore()
	generator := NewCertificateGenerator(store)

	caCert, err := generator.Generate(CertParams{
		CommonName: "registry-ca",
		IsCA:       true,
	})
	if err != nil {
		return fmt.Errorf("Generating CA: %s", err)
	}

	storedCAName := "ca"
	store.StoreCert(storedCAName, caCert)

	serverCert, err := generator.Generate(CertParams{
		CommonName:       ip,
		AlternativeNames: []string{ip}, // newer Golang needs this (eg kaniko)
		CAName:           storedCAName,
	})
	if err != nil {
		return fmt.Errorf("Generating CA: %s", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.tlsSecretName,
			Namespace: r.namespace,
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			corev1.TLSCertKey:       serverCert.Certificate,
			corev1.TLSPrivateKeyKey: serverCert.PrivateKey,
			tlsCACertKey:            serverCert.CA,
		},
	}

	_, err = r.coreClient.CoreV1().Secrets(r.namespace).Create(secret)
	if err != nil {
		return fmt.Errorf("Creating TLS secret: %s", err)
	}

	return nil
}

func (r Registry) getOrCreateAuthSecret() (string, string, error) {
	foundSecret, err := r.coreClient.CoreV1().Secrets(r.namespace).Get(r.authSecretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return "", "", fmt.Errorf("Getting auth secret: %s", err)
		}
		return r.createAuthSecret()
	}

	username := string(foundSecret.Data[corev1.BasicAuthUsernameKey])
	password := string(foundSecret.Data[corev1.BasicAuthPasswordKey])

	return username, password, nil
}

func (r Registry) createAuthSecret() (string, string, error) {
	const username = "admin"

	password, err := NewPasswordGenerator().Generate(PasswordParams{})
	if err != nil {
		return "", "", fmt.Errorf("Generating password: %s", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.authSecretName,
			Namespace: r.namespace,
		},
		Type: corev1.SecretTypeBasicAuth,
		StringData: map[string]string{
			corev1.BasicAuthUsernameKey: username,
			corev1.BasicAuthPasswordKey: password,
		},
	}

	_, err = r.coreClient.CoreV1().Secrets(r.namespace).Create(secret)
	if err != nil {
		return "", "", fmt.Errorf("Creating auth secret: %s", err)
	}

	return username, password, nil
}

func (r Registry) getOrCreateImagePullSecret(ip, username, password string) error {
	_, err := r.coreClient.CoreV1().Secrets(r.namespace).Get(r.dockerjsonSecretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("Getting image pull secret: %s", err)
		}
		return r.createImagePullSecret(ip, username, password)
	}
	return nil
}

func (r Registry) createImagePullSecret(ip, username, password string) error {
	content := map[string]interface{}{
		"auths": map[string]interface{}{
			"https://" + ip: map[string]interface{}{
				"username": username,
				"password": password,
				"email":    "noop",
			},
		},
	}

	contentBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.dockerjsonSecretName,
			Namespace: r.namespace,
			Annotations: map[string]string{
				"registry-address": ip,
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		StringData: map[string]string{
			corev1.DockerConfigJsonKey: string(contentBytes),
		},
	}

	_, err = r.coreClient.CoreV1().Secrets(r.namespace).Create(secret)
	if err != nil {
		return fmt.Errorf("Creating image pull secret: %s", err)
	}

	return nil
}

func (r Registry) waitForNsDeletion() error {
	timeoutCh := time.After(2 * time.Minute)

	for {
		_, err := r.coreClient.CoreV1().Namespaces().Get(r.namespace, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
		}

		select {
		case <-timeoutCh:
			return fmt.Errorf("Timed out waiting for namespace '%s' to be deleted", r.namespace)
		default:
			// continue with waiting
		}

		time.Sleep(1 * time.Second)
	}
}
