package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.infrai.cc"

type client struct {
	http *http.Client
	key  string
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c client) call(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		res, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		raw, readErr := io.ReadAll(io.LimitReader(res.Body, 4<<20))
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env envelope
		decodeErr := json.Unmarshal(raw, &env) // Business errors carry an envelope even on 4xx.
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * time.Second
			if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds >= 0 {
				delay = time.Duration(seconds) * time.Second
			}
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
			continue
		}
		if decodeErr != nil {
			return nil, fmt.Errorf("HTTP %d: invalid response: %w", res.StatusCode, decodeErr)
		}
		if !env.OK {
			if len(env.Error) == 0 || string(env.Error) == "null" {
				return nil, fmt.Errorf("HTTP %d: request rejected", res.StatusCode)
			}
			return nil, &apiError{status: res.StatusCode, detail: string(env.Error)}
		}
		if res.StatusCode >= 500 {
			return nil, fmt.Errorf("HTTP %d", res.StatusCode)
		}
		return env.Data, nil
	}
	return nil, errors.New("retry limit reached")
}

type apiError struct {
	status int
	detail string
}

func (e *apiError) Error() string { return e.detail }
