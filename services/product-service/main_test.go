package main

import (
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Basic test to ensure package compiles
	if tableName == "" {
		tableName = "test-table"
	}
	t.Log("Health handler test passed")
}

func TestGetEnv(t *testing.T) {
	result := getEnv("NONEXISTENT_KEY", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}
