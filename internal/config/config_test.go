package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	result := GetEnv("TEST_KEY", "default_value")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	result = GetEnv("MISSING_KEY", "default_value")
	if result != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	result := GetEnvInt("TEST_INT", 456)
	if result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}

	result = GetEnvInt("MISSING_INT", 456)
	if result != 456 {
		t.Errorf("Expected 456, got %d", result)
	}
}
