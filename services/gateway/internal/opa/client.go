package opa

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

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Input struct {
	AgentID     string `json:"agent_id"`
	AgentType   string `json:"agent_type"`
	SessionID   string `json:"session_id"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
	Environment string `json:"environment"`
	RiskScore   int    `json:"risk_score"`
}

type Decision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type requestBody struct {
	Input Input `json:"input"`
}

type responseBody struct {
	Result Decision `json:"result"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) Evaluate(
	ctx context.Context,
	input Input,
) (*Decision, error) {

	body, err := json.Marshal(requestBody{
		Input: input,
	})
	if err != nil {
		return nil, err
	}

	url := c.BaseURL + "/v1/data/agentshield/authz/decision"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OPA request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"OPA returned unexpected status: %s",
			resp.Status,
		)
	}

	var result responseBody

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Result.Decision == "" {
		return nil, errors.New("OPA returned empty decision")
	}

	return &result.Result, nil
}
