package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type Claims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func GenerateToken(secret string, ttl time.Duration, userID, email, name, role string) (string, error) {
	if secret == "" {
		return "", errors.New("jwt secret is required")
	}

	now := time.Now().UTC()
	hdrBytes, err := json.Marshal(header{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}

	claimsBytes, err := json.Marshal(Claims{
		Subject:   userID,
		Email:     email,
		Name:      name,
		Role:      role,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	signingInput := base64.RawURLEncoding.EncodeToString(hdrBytes) + "." + base64.RawURLEncoding.EncodeToString(claimsBytes)
	sig := signHS256(signingInput, secret)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func ParseToken(secret, tokenString string) (*Claims, error) {
	if secret == "" {
		return nil, ErrInvalidToken
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	hdrBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var hdr header
	if err := json.Unmarshal(hdrBytes, &hdr); err != nil || hdr.Alg != "HS256" || hdr.Typ != "JWT" {
		return nil, ErrInvalidToken
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}

	expectedSig := signHS256(parts[0]+"."+parts[1], secret)
	if !hmac.Equal(sigBytes, expectedSig) {
		return nil, ErrInvalidToken
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt <= time.Now().UTC().Unix() {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func signHS256(signingInput, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return mac.Sum(nil)
}
