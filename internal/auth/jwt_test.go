package auth

import (
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	token, err := manager.GenerateAccessToken(1, "admin", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("UserID = %d, want 1", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("Username = %v, want admin", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %v, want admin", claims.Role)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	token, err := manager.GenerateRefreshToken(2, "operator", "operator")
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if claims.UserID != 2 {
		t.Errorf("UserID = %d, want 2", claims.UserID)
	}
	if claims.Username != "operator" {
		t.Errorf("Username = %v, want operator", claims.Username)
	}
}

func TestInvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	_, err := manager.ValidateToken("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestWrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-1")
	manager2 := NewJWTManager("secret-2")

	token, _ := manager1.GenerateAccessToken(1, "admin", "admin")

	_, err := manager2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken with wrong secret, got %v", err)
	}
}

func TestEmptyToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	_, err := manager.ValidateToken("")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for empty token, got %v", err)
	}
}

func TestDifferentRoles(t *testing.T) {
	manager := NewJWTManager("test-secret")

	roles := []string{"admin", "operator", "viewer"}
	for _, role := range roles {
		token, err := manager.GenerateAccessToken(1, "user", role)
		if err != nil {
			t.Fatalf("Failed to generate token for role %s: %v", role, err)
		}

		claims, err := manager.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token for role %s: %v", role, err)
		}

		if claims.Role != role {
			t.Errorf("Role = %v, want %v", claims.Role, role)
		}
	}
}
