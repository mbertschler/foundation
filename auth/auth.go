package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Password hashing parameters
const (
	saltLength  = 32
	keyLength   = 32
	timeCost    = 3
	memoryCost  = 32 * 1024 // 32MB
	parallelism = 4
)

// HashPassword generates a hashed password using Argon2
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)

	// Encode salt and hash as base64
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$t=3,m=32768,p=4$salt$hash
	return fmt.Sprintf("$argon2id$t=%d,m=%d,p=%d$%s$%s",
		timeCost, memoryCost, parallelism, saltB64, hashB64), nil
}

// verifyPassword verifies a password against a hash
func verifyPassword(password, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		return false, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported hash type: %s", parts[1])
	}

	// Parse parameters from the hash string
	params := strings.Split(parts[2], ",")
	if len(params) != 3 {
		return false, fmt.Errorf("invalid parameters format")
	}

	var time, memory uint32
	var parallelism uint8
	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			return false, fmt.Errorf("invalid parameter format: %s", param)
		}
		switch kv[0] {
		case "t":
			if _, err := fmt.Sscanf(kv[1], "%d", &time); err != nil {
				return false, fmt.Errorf("invalid time parameter: %w", err)
			}
		case "m":
			if _, err := fmt.Sscanf(kv[1], "%d", &memory); err != nil {
				return false, fmt.Errorf("invalid memory parameter: %w", err)
			}
		case "p":
			var p uint32
			if _, err := fmt.Sscanf(kv[1], "%d", &p); err != nil {
				return false, fmt.Errorf("invalid parallelism parameter: %w", err)
			}
			parallelism = uint8(p)
		default:
			return false, fmt.Errorf("unknown parameter: %s", kv[0])
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, parallelism, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}
