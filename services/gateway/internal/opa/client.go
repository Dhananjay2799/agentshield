package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
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

type MatchedPolicy struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Effect   string `json:"effect"`
	Version  int    `json:"version"`
	Source   string `json:"source"`
}

type Decision struct {
	Decision      string         `json:"decision"`
	Reason        string         `json:"reason"`
	MatchedPolicy *MatchedPolicy `json:"matched_policy,omitempty"`
}

type requestBody struct {
	Input Input `json:"input"`
}

type responseBody struct {
	Result Decision `json:"result"`
}

type ManagedPolicy struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Effect        string  `json:"effect"`
	Priority      int     `json:"priority"`
	AgentType     *string `json:"agent_type,omitempty"`
	Action        string  `json:"action"`
	ActionMatch   string  `json:"action_match"`
	Resource      string  `json:"resource"`
	ResourceMatch string  `json:"resource_match"`
	Environment   *string `json:"environment,omitempty"`
	Version       int     `json:"version"`
	Source        string  `json:"source"`
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

	endpoint :=
		c.BaseURL +
			"/v1/data/agentshield/authz/decision"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil,
			fmt.Errorf(
				"OPA request failed: %w",
				err,
			)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil,
			fmt.Errorf(
				"OPA returned unexpected status: %s",
				resp.Status,
			)
	}

	var result responseBody

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&result); err != nil {
		return nil, err
	}

	if result.Result.Decision == "" {
		return nil,
			errors.New(
				"OPA returned empty decision",
			)
	}

	return &result.Result, nil
}

func (c *Client) PutManagedPolicy(
	ctx context.Context,
	policy *models.Policy,
) error {

	document := ManagedPolicy{
		ID:            policy.ID,
		Name:          policy.Name,
		Description:   policy.Description,
		Effect:        policy.Effect,
		Priority:      policy.Priority,
		AgentType:     policy.AgentType,
		Action:        policy.Action,
		ActionMatch:   policy.ActionMatch,
		Resource:      policy.Resource,
		ResourceMatch: policy.ResourceMatch,
		Environment:   policy.Environment,
		Version:       policy.Version,
		Source:        policy.Source,
	}

	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf(
			"marshal managed policy: %w",
			err,
		)
	}

	endpoint :=
		c.managedPolicyURL(policy.ID)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf(
			"create OPA managed policy request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"write managed policy to OPA: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {

		message, _ :=
			io.ReadAll(
				io.LimitReader(
					resp.Body,
					4096,
				),
			)

		return fmt.Errorf(
			"OPA managed policy write returned %s: %s",
			resp.Status,
			strings.TrimSpace(
				string(message),
			),
		)
	}

	return nil
}

func (c *Client) DeleteManagedPolicy(
	ctx context.Context,
	policyID string,
) error {

	endpoint :=
		c.managedPolicyURL(policyID)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		endpoint,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"create OPA managed policy delete request: %w",
			err,
		)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"delete managed policy from OPA: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {

		message, _ :=
			io.ReadAll(
				io.LimitReader(
					resp.Body,
					4096,
				),
			)

		return fmt.Errorf(
			"OPA managed policy delete returned %s: %s",
			resp.Status,
			strings.TrimSpace(
				string(message),
			),
		)
	}

	return nil
}

func (c *Client) managedPolicyURL(
	policyID string,
) string {
	return c.BaseURL +
		"/v1/data/agentshield_runtime/managed_policies/" +
		url.PathEscape(policyID)
}
