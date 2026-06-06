package auth

import (
	"os"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	os.Setenv("SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("SECRET_KEY")

	service := NewService()
	token, err := service.GenerateToken(1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestValidateToken(t *testing.T) {
	os.Setenv("SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("SECRET_KEY")

	service := NewService()
	token, err := service.GenerateToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	validatedToken, err := service.ValidateToken(token)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !validatedToken.Valid {
		t.Error("Expected valid token")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	os.Setenv("SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("SECRET_KEY")

	service := NewService()
	_, err := service.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	os.Setenv("SECRET_KEY", "secret-a")
	serviceA := NewService()
	token, _ := serviceA.GenerateToken(1)
	os.Unsetenv("SECRET_KEY")

	os.Setenv("SECRET_KEY", "secret-b")
	serviceB := NewService()
	validatedToken, err := serviceB.ValidateToken(token)
	os.Unsetenv("SECRET_KEY")

	if err == nil && validatedToken.Valid {
		t.Error("Token verified with wrong secret should be invalid")
	}
}
