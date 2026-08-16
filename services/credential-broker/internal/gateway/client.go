package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AuthorizationGrant struct {
	ID         string     `json:"id"`
	ApprovalID string     `json:"approval_id"`
	AgentID    string     `json:"agent_id"`
	SessionID  string     `json:"session_id"`
	Action     string     `json:"action"`
	Resource   string     `json:"resource"`
	Status     string     `json:"status"`
	IssuedAt   time.Time  `json:"issued_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
}

type GrantVerification struct {
	Valid  bool               `json:"valid"`
	Reason string             `json:"reason"`
	Grant  AuthorizationGrant `json:"grant"`
}

type GrantClaim struct {
	Claimed bool               `json:"claimed"`
	Grant   AuthorizationGrant `json:"grant"`
	Error   string             `json:"error,omitempty"`
}

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(
	baseURL string,
	apiKey string,
) *Client {
	return &Client{
		BaseURL: strings.TrimRight(
			baseURL,
			"/",
		),
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) VerifyGrant(
	ctx context.Context,
	grantID string,
) (*GrantVerification, error) {
	if strings.TrimSpace(grantID) == "" {
		return nil, errors.New(
			"grant id is required",
		)
	}

	url := fmt.Sprintf(
		"%s/v1/grants/%s/verify",
		c.BaseURL,
		grantID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Accept",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+c.APIKey,
	)

	response, err :=
		c.HTTPClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode ==
		http.StatusNotFound {
		return &GrantVerification{
			Valid:  false,
			Reason: "authorization grant not found",
		}, nil
	}

	if response.StatusCode !=
		http.StatusOK {
		return nil, fmt.Errorf(
			"gateway grant verification returned status %d",
			response.StatusCode,
		)
	}

	var verification GrantVerification

	if err := json.NewDecoder(
		response.Body,
	).Decode(
		&verification,
	); err != nil {
		return nil, err
	}

	return &verification, nil
}

func (c *Client) ClaimGrant(
	ctx context.Context,
	grantID string,
	agentID string,
	sessionID string,
	action string,
	resource string,
) (*GrantClaim, error) {
	if strings.TrimSpace(grantID) == "" {
		return nil, errors.New("grant id is required")
	}

	body := map[string]string{
		"agent_id":   agentID,
		"session_id": sessionID,
		"action":     action,
		"resource":   resource,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"%s/v1/grants/%s/claim",
		c.BaseURL,
		grantID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(bodyJSON),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Accept",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+c.APIKey,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	response, err :=
		c.HTTPClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var claim GrantClaim

	if err := json.NewDecoder(
		response.Body,
	).Decode(&claim); err != nil {
		return nil, err
	}

	if response.StatusCode ==
		http.StatusConflict {
		claim.Claimed = false
		return &claim, nil
	}

	if response.StatusCode !=
		http.StatusOK {
		return nil, fmt.Errorf(
			"gateway grant claim returned status %d",
			response.StatusCode,
		)
	}

	return &claim, nil
}
