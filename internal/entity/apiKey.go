package entity

import (
	"crypto/rand"
	"encoding/hex"
)

const (
	ApiKeyRateKey       = "rate:api-key"
	ApiKeyBlockDuration = "block:api-key"
	StatusApiKeyBlock   = "ApiKeyBlock"
	ApiKeyHeader        = "API_KEY"
)

type ApiKey struct {
	value         string
	BlockDuration int64
	RateLimiter   RateLimiter
}

func (ap *ApiKey) GenerateValue() error {
	bytes, err := generateRandomBytes(32)
	if err != nil {
		return err
	}
	ap.value = hex.EncodeToString(bytes)
	return nil
}

func generateRandomBytes(length int) ([]byte, error) {
	byteSlice := make([]byte, length)
	_, err := rand.Read(byteSlice)
	if err != nil {
		return nil, err
	}

	return byteSlice, nil
}

func (ap *ApiKey) SetValue(value string) {
	ap.value = value
}

func (ap *ApiKey) Value() string {
	return ap.value
}

func (ap *ApiKey) Validate() error {
	if ap.BlockDuration == 0 {
		return ErrBlockTimeDuration
	}

	if err := ap.RateLimiter.Validate(); err != nil {
		return err
	}

	return nil
}
