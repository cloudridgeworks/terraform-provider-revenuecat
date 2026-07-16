// Copyright Cloud Ridge Works 2026
// SPDX-License-Identifier: MPL-2.0

package client

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
)

const DefaultBaseURL = "https://api.revenuecat.com/v2"

var ErrNotFound = errors.New("RevenueCat resource not found")

type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("RevenueCat API returned HTTP %d: %s", e.StatusCode, e.Body)
}

func New(apiKey, baseURL string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("RevenueCat API key must not be empty")
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse RevenueCat base URL: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1" {
		return nil, errors.New("RevenueCat base URL must use HTTPS")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{baseURL: parsed, apiKey: apiKey, httpClient: httpClient}, nil
}

func (c *Client) Do(ctx context.Context, method, path string, body, result any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode RevenueCat request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL.String()+"/"+strings.TrimLeft(path, "/"), reader)
	if err != nil {
		return fmt.Errorf("create RevenueCat request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-revenuecat")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call RevenueCat API: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return fmt.Errorf("read RevenueCat response: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(payload))}
	}
	if result != nil && len(payload) != 0 {
		if err := json.Unmarshal(payload, result); err != nil {
			return fmt.Errorf("decode RevenueCat response: %w", err)
		}
	}
	return nil
}

type ProductAssociationList struct {
	Items []struct {
		ID                  string `json:"id"`
		EligibilityCriteria string `json:"eligibility_criteria"`
		Product             struct {
			ID string `json:"id"`
		} `json:"product"`
	} `json:"items"`
}

func (c *Client) FindProductAssociation(ctx context.Context, path, productID string) (string, bool, error) {
	var list ProductAssociationList
	if err := c.Do(ctx, http.MethodGet, path+"?limit=100", nil, &list); err != nil {
		return "", false, err
	}
	for _, item := range list.Items {
		id := item.ID
		if item.Product.ID != "" {
			id = item.Product.ID
		}
		if id == productID {
			return item.EligibilityCriteria, true, nil
		}
	}
	return "", false, nil
}
