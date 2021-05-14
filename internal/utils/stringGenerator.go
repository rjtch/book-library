package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strconv"
)

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int64) (string, error) {
	h := sha256.New()
	io.WriteString(h, strconv.FormatInt(s, 10))
	io.WriteString(h, "ganraomaxxxxxxxxx")
	token := fmt.Sprintf("%x", h.Sum(nil))
	return token, nil
}
