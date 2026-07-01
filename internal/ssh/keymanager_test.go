package ssh

import (
	"strings"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	km := NewKeyManager()

	privateKey, publicKey, fingerprint, err := km.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if privateKey == "" {
		t.Error("Private key is empty")
	}
	if !strings.HasPrefix(publicKey, "ssh-rsa") {
		t.Errorf("Public key doesn't start with ssh-rsa: %v", publicKey)
	}
	if fingerprint == "" {
		t.Error("Fingerprint is empty")
	}
	if !strings.HasPrefix(fingerprint, "SHA256:") {
		t.Errorf("Fingerprint doesn't start with SHA256: %v", fingerprint)
	}
}

func TestGetFingerprint(t *testing.T) {
	km := NewKeyManager()

	privateKey, _, fingerprint1, _ := km.GenerateKeyPair(2048)

	fingerprint2, err := km.GetFingerprint(privateKey)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}

	if fingerprint1 != fingerprint2 {
		t.Errorf("Fingerprints don't match: %v != %v", fingerprint1, fingerprint2)
	}
}

func TestInvalidPrivateKey(t *testing.T) {
	km := NewKeyManager()

	_, err := km.GetFingerprint("not-a-real-key")
	if err == nil {
		t.Error("Expected error for invalid private key")
	}
}

func TestMultipleKeyPairs(t *testing.T) {
	km := NewKeyManager()

	fp1, fp2 := "", ""
	for i := 0; i < 3; i++ {
		_, _, fp, err := km.GenerateKeyPair(2048)
		if err != nil {
			t.Fatalf("Failed to generate key pair %d: %v", i, err)
		}
		if i == 0 {
			fp1 = fp
		}
		if i == 1 {
			fp2 = fp
		}
	}

	if fp1 == fp2 {
		t.Errorf("Two different key pairs should have different fingerprints")
	}
}
