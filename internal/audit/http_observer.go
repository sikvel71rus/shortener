package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type HTTPObserver struct {
	target string
	client *http.Client
}

// NewHTTPObserver creates an observer that posts audit events to a remote endpoint.
func NewHTTPObserver(target string) (*HTTPObserver, error) {
	if target == "" {
		return nil, nil
	}

	if _, err := url.ParseRequestURI(target); err != nil {
		return nil, fmt.Errorf("invalid audit url: %w", err)
	}

	return &HTTPObserver{
		target: target,
		client: &http.Client{},
	}, nil
}

// Notify sends the audit event to the configured HTTP endpoint.
func (o *HTTPObserver) Notify(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.target, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected audit status: %s", resp.Status)
	}

	return nil
}
