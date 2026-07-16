// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestClientDo(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		if r.URL.Path != "/v2/projects/proj/entitlements/ent" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"id":"ent"}`)), Header: make(http.Header)}, nil
	})}

	c, err := New("secret", "https://api.example.com/v2", httpClient)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := c.Do(context.Background(), http.MethodGet, "projects/proj/entitlements/ent", nil, &response); err != nil {
		t.Fatal(err)
	}
	if response.ID != "ent" {
		t.Fatalf("unexpected response ID %q", response.ID)
	}
}

func TestClientNotFound(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("missing")), Header: make(http.Header)}, nil
	})}
	c, err := New("secret", "https://api.example.com/v2", httpClient)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Do(context.Background(), http.MethodGet, "missing", nil, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestClientRejectsInsecureRemoteURL(t *testing.T) {
	if _, err := New("secret", "http://api.example.com/v2", nil); err == nil {
		t.Fatal("expected insecure URL to be rejected")
	}
}
