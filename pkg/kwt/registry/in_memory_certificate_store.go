package registry

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

type InMemoryCertificateStore struct {
	store map[string]CertResponse
}

var _ CertsLoader = &InMemoryCertificateStore{}

func NewInMemoryCertificateStore() *InMemoryCertificateStore {
	return &InMemoryCertificateStore{map[string]CertResponse{}}
}

func (s *InMemoryCertificateStore) StoreCert(name string, resp CertResponse) {
	s.store[name] = resp
}

func (s *InMemoryCertificateStore) LoadCerts(name string) (*x509.Certificate, *rsa.PrivateKey, error) {
	resp, found := s.store[name]
	if !found {
		return nil, nil, fmt.Errorf("Expected to find cert '%s' but did not", name)
	}

	cpb, _ := pem.Decode([]byte(resp.Certificate))

	crt, err := x509.ParseCertificate(cpb.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("Parsing certificate: %s", err)
	}

	kpb, _ := pem.Decode([]byte(resp.PrivateKey))

	key, err := x509.ParsePKCS1PrivateKey(kpb.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("Parsing private key: %s", err)
	}

	return crt, key, nil
}
