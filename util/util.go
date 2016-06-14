package util

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"

	"github.com/pkg/errors"
)

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)

	if err != nil {
		return "", errors.Wrap(err, "Failed to generate random strings.")
	}

	return strings.TrimRight(base64.StdEncoding.EncodeToString(b), "="), nil
}
