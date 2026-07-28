// Package auth contains the cryptographic primitives for JOBHOO
// authentication: password hashing and session token generation/hashing.
// Handlers and middleware call into this package; they never call bcrypt or
// crypto/rand directly, so there is exactly one place that decides how
// passwords and tokens are handled.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plaintext password for storage. bcrypt's cost factor
// of 12 is a deliberate balance: slow enough to resist offline brute force,
// fast enough not to bottleneck login under normal load.
func HashPassword(plaintext string) (string, error) {
	if len(plaintext) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword reports whether plaintext matches the stored bcrypt hash.
func VerifyPassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// NewSessionToken generates a cryptographically random session token
// suitable for use as a cookie value, along with the SHA-256 hash that gets
// stored in the database. Only the hash is ever persisted.
func NewSessionToken() (rawToken string, tokenHash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	rawToken = base64.RawURLEncoding.EncodeToString(buf)
	tokenHash = HashToken(rawToken)
	return rawToken, tokenHash, nil
}

// HashToken deterministically hashes a raw session token for lookup/storage.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
