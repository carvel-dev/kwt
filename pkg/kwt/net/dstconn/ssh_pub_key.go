package dstconn

import (
	"encoding/base64"
	"fmt"
	"strings"

	gossh "golang.org/x/crypto/ssh"
)

// Opposite of gossh.MarshalAuthorizedKey
func ParsePublicKey(pubKey string) (gossh.PublicKey, error) {
	pieces := strings.SplitN(strings.TrimSpace(pubKey), " ", 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("Expected public key to have type and content")
	}

	bs, err := base64.StdEncoding.DecodeString(pieces[1])
	if err != nil {
		return nil, fmt.Errorf("Decoding public key content: %s", err)
	}

	hostPublicKey, err := gossh.ParsePublicKey(bs)
	if err != nil {
		return nil, fmt.Errorf("Parsing host public key: %s", err)
	}

	return hostPublicKey, nil
}
