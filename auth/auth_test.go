package auth

import (
	"strings"
	"testing"
)

func TesthashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword failed: %v", err)
	}

	// Check that hash is not empty
	if hash == "" {
		t.Error("hashPassword returned empty string")
	}

	// Check hash format
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		t.Errorf("Invalid hash format, expected 5 parts, got %d", len(parts))
	}

	if parts[1] != "argon2id" {
		t.Errorf("Expected argon2id, got %s", parts[1])
	}

	// Check parameters format
	params := strings.Split(parts[2], ",")
	if len(params) != 3 {
		t.Errorf("Invalid parameters format, expected 3 params, got %d", len(params))
	}

	// Verify parameters contain expected keys
	paramMap := make(map[string]bool)
	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) == 2 {
			paramMap[kv[0]] = true
		}
	}

	if !paramMap["t"] || !paramMap["m"] || !paramMap["p"] {
		t.Error("Missing expected parameters t, m, or p")
	}

	// Check salt and hash are base64 encoded
	if parts[3] == "" {
		t.Error("Salt is empty")
	}
	if parts[4] == "" {
		t.Error("Hash is empty")
	}
}

func TestverifyPassword(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword failed: %v", err)
	}

	// Test correct password
	valid, err := verifyPassword(password, hash)
	if err != nil {
		t.Fatalf("verifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("verifyPassword should return true for correct password")
	}

	// Test wrong password
	valid, err = verifyPassword(wrongPassword, hash)
	if err != nil {
		t.Fatalf("verifyPassword failed: %v", err)
	}
	if valid {
		t.Error("verifyPassword should return false for wrong password")
	}
}

func TestverifyPasswordInvalidFormats(t *testing.T) {
	password := "testpassword"

	testCases := []struct {
		name string
		hash string
	}{
		{"empty hash", ""},
		{"too few parts", "$argon2id$t=3,m=32768,p=4$salt"},
		{"too many parts", "$argon2id$t=3,m=32768,p=4$salt$hash$extra"},
		{"wrong algorithm", "$argon2i$t=3,m=32768,p=4$salt$hash"},
		{"invalid parameters", "$argon2id$invalid$salt$hash"},
		{"missing parameter", "$argon2id$t=3,m=32768$salt$hash"},
		{"invalid parameter format", "$argon2id$t=3,m=32768,p=4,invalid$salt$hash"},
		{"invalid salt base64", "$argon2id$t=3,m=32768,p=4$invalidbase64$hash"},
		{"invalid hash base64", "$argon2id$t=3,m=32768,p=4$salt$invalidbase64"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := verifyPassword(password, tc.hash)
			if err == nil {
				t.Errorf("Expected error for invalid hash format: %s", tc.name)
			}
			if valid {
				t.Errorf("Should return false for invalid hash: %s", tc.name)
			}
		})
	}
}

func TestPasswordHashingConsistency(t *testing.T) {
	password := "consistentpassword"

	// Generate multiple hashes for the same password
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First hashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second hashPassword failed: %v", err)
	}

	// Hashes should be different due to different salts
	if hash1 == hash2 {
		t.Error("Hashes should be different due to random salts")
	}

	// But both should verify the same password
	valid1, err := verifyPassword(password, hash1)
	if err != nil {
		t.Fatalf("verifyPassword for hash1 failed: %v", err)
	}
	if !valid1 {
		t.Error("hash1 should verify the password")
	}

	valid2, err := verifyPassword(password, hash2)
	if err != nil {
		t.Fatalf("verifyPassword for hash2 failed: %v", err)
	}
	if !valid2 {
		t.Error("hash2 should verify the password")
	}
}

func TestEdgeCases(t *testing.T) {
	// Test empty password
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("hashPassword failed for empty password: %v", err)
	}

	valid, err := verifyPassword("", hash)
	if err != nil {
		t.Fatalf("verifyPassword failed for empty password: %v", err)
	}
	if !valid {
		t.Error("Empty password should verify correctly")
	}

	// Test long password
	longPassword := strings.Repeat("a", 1000)
	hash, err = HashPassword(longPassword)
	if err != nil {
		t.Fatalf("hashPassword failed for long password: %v", err)
	}

	valid, err = verifyPassword(longPassword, hash)
	if err != nil {
		t.Fatalf("verifyPassword failed for long password: %v", err)
	}
	if !valid {
		t.Error("Long password should verify correctly")
	}
}
