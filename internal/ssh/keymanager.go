package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

type KeyManager struct{}

func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

func (km *KeyManager) GenerateKeyPair(bits int) (privateKey string, publicKey string, fingerprint string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Marshal private key to PEM (PKCS#1 format)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	// Generate public key
	pubKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(pubKey)

	// Calculate fingerprint (SHA256)
	hash := sha256.Sum256(pubKey.Marshal())
	fingerprint = "SHA256:" + hex.EncodeToString(hash[:])

	return string(privateKeyPEM), string(publicKeyBytes), fingerprint, nil
}

func (km *KeyManager) GetFingerprint(privateKey string) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	pubKey := signer.PublicKey()
	hash := sha256.Sum256(pubKey.Marshal())
	return "SHA256:" + hex.EncodeToString(hash[:]), nil
}
