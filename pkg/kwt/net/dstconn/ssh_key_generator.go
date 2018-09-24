package dstconn

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

const (
	sshKeyGeneratorKeyBits          = 4096
	sshKeyGeneratorHeaderPrivateKey = "RSA PRIVATE KEY"
)

type SSHKeyGenerator struct{}

type SSHKey struct {
	PrivateKey string
	PublicKey  string
}

func NewSSHKeyGenerator() SSHKeyGenerator {
	return SSHKeyGenerator{}
}

func (g SSHKeyGenerator) Generate() (SSHKey, error) {
	priv, pub, err := g.generateRSAKeyPair()
	if err != nil {
		return SSHKey{}, fmt.Errorf("Generating RSA key pair: %s", err)
	}

	sshPubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return SSHKey{}, err
	}

	key := SSHKey{
		PrivateKey: g.privateKeyToPEM(priv),
		PublicKey:  string(ssh.MarshalAuthorizedKey(sshPubKey)),
	}

	return key, nil
}

func (g SSHKeyGenerator) encodePEM(keyBytes []byte, keyType string) string {
	block := &pem.Block{
		Type:  keyType,
		Bytes: keyBytes,
	}
	return string(pem.EncodeToMemory(block))
}

func (g SSHKeyGenerator) generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, sshKeyGeneratorKeyBits)
	if err != nil {
		return nil, nil, err
	}
	public := private.Public().(*rsa.PublicKey)
	return private, public, nil
}

func (g SSHKeyGenerator) privateKeyToPEM(privateKey *rsa.PrivateKey) string {
	return g.encodePEM(x509.MarshalPKCS1PrivateKey(privateKey), sshKeyGeneratorHeaderPrivateKey)
}
