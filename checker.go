package main

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Result describes the outcome of one health check.
type Result struct {
	Up         bool
	StatusCode int
	Duration   time.Duration
	Err        error
}

// Check sends a GET request to target and reports whether it is up.
// A target is "up" when it answers with a status code in the 2xx or 3xx range.
// The deadline comes from ctx, so the caller controls the timeout.
func Check(ctx context.Context, client *http.Client, target string) Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return Result{Err: err}
	}

	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return Result{Duration: elapsed, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	return Result{
		Up:         resp.StatusCode >= 200 && resp.StatusCode < 400,
		StatusCode: resp.StatusCode,
		Duration:   elapsed,
	}
}
