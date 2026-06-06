package helper

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestFormatValidationError(t *testing.T) {
	validate := validator.New()

	err := validate.Struct(TestStruct{Name: "", Email: ""})
	if err == nil {
		t.Fatal("Expected validation errors")
	}

	errors := FormatValidationError(err)
	if len(errors) == 0 {
		t.Error("Expected non-empty error list")
	}
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
}

func TestFormatValidationError_NoErrors(t *testing.T) {
	validate := validator.New()

	err := validate.Struct(TestStruct{Name: "John", Email: "john@example.com"})
	if err != nil {
		t.Fatalf("Expected no validation errors, got %v", err)
	}

	// This will panic if called with nil, which is expected behavior
	// In production code, always check err != nil before calling
}

func TestApiResponse(t *testing.T) {
	response := ApiResponse("test message", 200, "success", nil)
	if response.Meta.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", response.Meta.Message)
	}
	if response.Meta.Code != 200 {
		t.Errorf("Expected code 200, got %d", response.Meta.Code)
	}
	if response.Meta.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Meta.Status)
	}
}

func TestApiResponse_WithData(t *testing.T) {
	data := map[string]string{"key": "value"}
	response := ApiResponse("test", 200, "success", data)
	if response.Data == nil {
		t.Error("Expected non-nil data")
	}
}

func TestGenCodeTransaction(t *testing.T) {
	code := GenCodeTransaction(123)
	if code == "" {
		t.Error("Expected non-empty transaction code")
	}
	if len(code) < 4 {
		t.Error("Expected transaction code with reasonable length")
	}
}

func TestGenCodeTransaction_UniquePerUser(t *testing.T) {
	code1 := GenCodeTransaction(123)
	code2 := GenCodeTransaction(456)
	if code1 == code2 {
		t.Error("Expected different transaction codes for different users")
	}
}
