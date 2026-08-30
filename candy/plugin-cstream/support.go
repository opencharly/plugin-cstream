package cstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// httpGet fetches a URL and returns the body, bounded so a hung gateway fails the
// step instead of hanging the run.
func httpGet(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	// Bounded read: /healthz is a small JSON object, and an unbounded read off a
	// misbehaving endpoint is a way to hang the check runner.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return string(body), fmt.Errorf("gateway returned %s", resp.Status)
	}
	return string(body), nil
}
