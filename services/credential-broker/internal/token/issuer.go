package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	CredentialID string `json:"credential_id"`
	GrantID      string `json:"grant_id"`
	AgentID      string `json:"agent_id"`
	SessionID    string `json:"session_id"`
	Action       string `json:"action"`
	Resource     string `json:"resource"`
	IssuedAt     int64  `json:"iat"`
	ExpiresAt    int64  `json:"exp"`
}

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(
	secret string,
	ttl time.Duration,
) (*Issuer, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("credential signing secret is required")
	}

	if ttl <= 0 {
		return nil, errors.New("credential ttl must be positive")
	}

	return &Issuer{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

func (i *Issuer) Issue(
	grantID string,
	agentID string,
	sessionID string,
	action string,
	resource string,
) (string, Claims, error) {
	credentialID, err := randomID()
	if err != nil {
		return "", Claims{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(i.ttl)

	claims := Claims{
		CredentialID: credentialID,
		GrantID:      grantID,
		AgentID:      agentID,
		SessionID:    sessionID,
		Action:       action,
		Resource:     resource,
		IssuedAt:     now.Unix(),
		ExpiresAt:    expiresAt.Unix(),
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", Claims{}, err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", Claims{}, err
	}

	encoder := base64.RawURLEncoding

	headerPart := encoder.EncodeToString(headerJSON)
	payloadPart := encoder.EncodeToString(claimsJSON)

	unsignedToken := headerPart + "." + payloadPart

	mac := hmac.New(
		sha256.New,
		i.secret,
	)

	_, _ = mac.Write(
		[]byte(unsignedToken),
	)

	signature := encoder.EncodeToString(
		mac.Sum(nil),
	)

	return fmt.Sprintf(
		"%s.%s",
		unsignedToken,
		signature,
	), claims, nil
}

func randomID() (
	string,
	error,
) {
	bytes := make(
		[]byte,
		16,
	)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
