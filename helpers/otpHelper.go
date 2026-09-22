package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// OTP lifetimes and limits — keep them in one place so they're easy to tune.
const (
	OTPLength     = 6
	OTPValidFor   = 10 * time.Minute
	OTPMaxAttempts = 5
)

// GenerateOTP returns a random 6-digit numeric code as a zero-padded
// string (e.g. "004271"), plus its SHA-256 hash for storage.
//
// We use crypto/rand (not math/rand) so codes are cryptographically
// unpredictable. math/rand would let an attacker guess future codes
// by observing past ones.
func GenerateOTP() (raw string, hash string, err error) {
	// 10^6 = 1,000,000, giving us a range of 000000–999999.
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", "", err
	}
	// %06d zero-pads to 6 digits: 4271 → "004271".
	raw = fmt.Sprintf("%06d", n.Int64())
	return raw, HashOTP(raw), nil
}

// HashOTP returns the hex SHA-256 hash of an OTP.
//
// Why not bcrypt? A 6-digit code has only 1M possibilities — bcrypt
// would slow down legitimate verification without meaningfully slowing
// an attacker who's already constrained to 5 tries. SHA-256 is enough
// because the DB field is a one-way digest, and the attempt cap makes
// brute force impossible regardless of hash speed.
func HashOTP(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}