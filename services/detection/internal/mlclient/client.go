package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(
	baseURL string,
	timeout time.Duration,
) *Client {
	baseURL = strings.TrimRight(
		baseURL,
		"/",
	)

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Predict(
	ctx context.Context,
	input PredictionRequest,
) (*PredictionResponse, error) {
	if c == nil {
		return nil, errors.New(
			"ML anomaly client is nil",
		)
	}

	if c.baseURL == "" {
		return nil, errors.New(
			"ML anomaly service URL is empty",
		)
	}

	payload, err :=
		json.Marshal(input)

	if err != nil {
		return nil, fmt.Errorf(
			"marshal ML prediction request: %w",
			err,
		)
	}

	request, err :=
		http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.baseURL+"/predict",
			bytes.NewReader(payload),
		)

	if err != nil {
		return nil, fmt.Errorf(
			"create ML prediction request: %w",
			err,
		)
	}

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response, err :=
		c.httpClient.Do(request)

	if err != nil {
		return nil, fmt.Errorf(
			"call ML anomaly service: %w",
			err,
		)
	}

	defer response.Body.Close()

	body, err :=
		io.ReadAll(
			io.LimitReader(
				response.Body,
				1<<20,
			),
		)

	if err != nil {
		return nil, fmt.Errorf(
			"read ML prediction response: %w",
			err,
		)
	}

	if response.StatusCode < 200 ||
		response.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"ML anomaly service returned status %d: %s",
			response.StatusCode,
			string(body),
		)
	}

	var prediction PredictionResponse

	if err :=
		json.Unmarshal(
			body,
			&prediction,
		); err != nil {
		return nil, fmt.Errorf(
			"decode ML prediction response: %w",
			err,
		)
	}

	return &prediction, nil
}
