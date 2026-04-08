package main

import (
	"strings"
	"testing"
)

func TestEnsureColimaReturnsInstallGuidance(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := ensureColima()
	if err == nil {
		t.Fatal("ensureColima() returned nil error")
	}
	if !strings.Contains(err.Error(), "brew install colima") {
		t.Fatalf("unexpected error: %v", err)
	}
}
