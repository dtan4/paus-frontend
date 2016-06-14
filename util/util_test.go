package util

import (
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	s, err := GenerateRandomString()

	if err != nil {
		t.Fatalf("Error should not be raised. error %s", err)
	}

	if len(s) != 43 {
		t.Fatalf("Generated string should have 43 characters. result: %s (%d chars)", s, len(s))
	}
}
