package signing

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"os"

	"k8s.io/apimachinery/pkg/util/json"
)

var privateKey string
var publicKey string

func init() {
	privateKey = os.Getenv("TEST_SIGN_KEY")
	publicKey = os.Getenv("TEST_READ_KEY")
}

func Sign(obj any) ([]byte, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object: %w", err)
	}

	decodedKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	b, _ := pem.Decode(decodedKey)

	key, err := x509.ParsePKIXPublicKey(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("failed to parse public key, must be RSA public key")
	}

	sum256Hash := sha256.Sum256(bytes)
	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, sum256Hash[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func Verify(obj any, signature []byte) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}

	decodedKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return fmt.Errorf("failed to decode private key: %w", err)
	}

	b, _ := pem.Decode(decodedKey)

	key, err := x509.ParsePKIXPublicKey(b.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("failed to parse public key, must be RSA public key")
	}

	sum256Hash := sha256.Sum256(bytes)
	err = rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, sum256Hash[:], signature)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}
	return nil
}
