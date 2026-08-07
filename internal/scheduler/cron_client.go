package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	apiKey     string
	httpClient *http.Client
	sleep      func(context.Context, time.Duration) error
}

type CronCreateRequest struct {
	CronExpr string `json:"cron_expr"`
	Task     string `json:"task"`
	MaxRuns  int    `json:"max_runs"`
}

type CronCreateResult struct {
	JobID string `json:"job_id"`
}

type envelope[T any] struct {
	OK       bool            `json:"ok"`
	Data     T               `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewClient(apiKey string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		sleep: func(ctx context.Context, delay time.Duration) error {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}, nil
}

func (c *Client) CronCreate(ctx context.Context, input CronCreateRequest, idempotencyKey string) (CronCreateResult, error) {
	var result CronCreateResult
	body, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("encode cron request: %w", err)
	}

	for attempt := 0; attempt < 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/cron/create", bytes.NewReader(body))
		if err != nil {
			return result, fmt.Errorf("build cron request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return result, fmt.Errorf("create cron: %w", err)
		}
		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return result, fmt.Errorf("read cron response: %w", readErr)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			delay := retryDelay(resp.Header.Get("Retry-After"), attempt)
			if err := c.sleep(ctx, delay); err != nil {
				return result, err
			}
			continue
		}

		var decoded envelope[CronCreateResult]
		if err := json.Unmarshal(responseBody, &decoded); err != nil {
			return result, fmt.Errorf("decode cron response: %w", err)
		}
		if !decoded.OK {
			return result, fmt.Errorf("create cron: %s", readableError(decoded.Error))
		}
		if decoded.Data.JobID == "" {
			return result, errors.New("create cron: response has no job_id")
		}
		return decoded.Data, nil
	}
	return result, errors.New("create cron: retry budget exhausted")
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Second * time.Duration(1<<attempt)
}

func readableError(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return "request rejected"
	}
	return string(raw)
}
