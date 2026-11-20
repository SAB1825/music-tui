// Package youtube discovers public catalog metadata through YouTube Data API v3.
package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/SAB1825/music-tui/internal/providers"
)

const Endpoint = "https://www.googleapis.com/youtube/v3"
const responseLimit = 4 << 20

type Client struct {
	key, endpoint string
	http          *http.Client
	timeout       time.Duration
	limit         int
}

func New(key, endpoint string, transport http.RoundTripper, timeout time.Duration, limit int) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if endpoint == "" {
		endpoint = Endpoint
	}
	return &Client{key: key, endpoint: strings.TrimRight(endpoint, "/"), timeout: timeout, limit: min(50, max(1, limit)),
		http: &http.Client{Transport: transport, Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (c *Client) get(ctx context.Context, operation string, query url.Values, out any) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	for attempt := range 3 {
		failure, delay := c.request(ctx, operation, query, out)
		if failure == nil {
			return nil
		}
		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return ctx.Err()
			}
			return &providers.Error{Kind: providers.Timeout, Operation: operation}
		}
		if attempt == 2 || delay > 2*time.Second || !retryable(failure.Kind) {
			return failure
		}
		delay = max(delay, 150*time.Millisecond<<attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			if errors.Is(ctx.Err(), context.Canceled) {
				return ctx.Err()
			}
			return &providers.Error{Kind: providers.Timeout, Operation: operation}
		case <-timer.C:
		}
	}
	return &providers.Error{Kind: providers.Service, Operation: operation}
}

func retryable(kind providers.ErrorKind) bool {
	return kind == providers.Network || kind == providers.RateLimit || kind == providers.Service
}

func (c *Client) request(ctx context.Context, operation string, query url.Values, out any) (*providers.Error, time.Duration) {
	fail := &providers.Error{Operation: operation}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/"+operation+"?"+query.Encode(), nil)
	if err != nil {
		fail.Kind = providers.InvalidInput
		return fail, 0
	}
	req.Header.Set("X-Goog-Api-Key", c.key)
	resp, err := c.http.Do(req)
	if err != nil {
		fail.Kind = providers.Network
		if e, ok := errors.AsType[net.Error](err); ok && e.Timeout() {
			fail.Kind = providers.Timeout
		}
		return fail, 0
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, responseLimit+1))
	closeErr := resp.Body.Close()
	fail.Status = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		fail.Kind = classify(resp.StatusCode, body)
		return fail, retryAfter(resp.Header.Get("Retry-After"))
	}
	if readErr != nil || closeErr != nil {
		fail.Kind = providers.Network
		return fail, 0
	}
	if len(body) > responseLimit || !bytes.HasPrefix(bytes.TrimSpace(body), []byte("{")) || json.Unmarshal(body, out) != nil {
		fail.Kind = providers.InvalidResponse
		return fail, 0
	}
	return nil, 0
}

func classify(status int, body []byte) providers.ErrorKind {
	var failure struct {
		Error struct {
			Errors  []struct{ Reason string }
			Details []struct{ Reason string }
		}
	}
	if json.Unmarshal(body, &failure) == nil {
		for _, reasons := range [][]struct{ Reason string }{failure.Error.Errors, failure.Error.Details} {
			for _, detail := range reasons {
				switch detail.Reason {
				case "quotaExceeded", "dailyLimitExceeded", "dailyLimitExceededUnreg":
					return providers.Quota
				case "keyInvalid", "API_KEY_INVALID", "API_KEY_SERVICE_BLOCKED", "API_KEY_HTTP_REFERRER_BLOCKED", "API_KEY_IP_ADDRESS_BLOCKED", "ipRefererBlocked":
					return providers.Credentials
				case "accessNotConfigured", "SERVICE_DISABLED":
					return providers.Disabled
				case "rateLimitExceeded", "userRateLimitExceeded":
					return providers.RateLimit
				}
			}
		}
	}
	switch status {
	case 401:
		return providers.Credentials
	case 403, 404, 410:
		return providers.Unavailable
	case 429:
		return providers.RateLimit
	case 500, 502, 503, 504:
		return providers.Service
	default:
		return providers.InvalidInput
	}
}

func retryAfter(value string) time.Duration {
	if n, err := strconv.Atoi(value); err == nil && n >= 0 {
		return time.Duration(min(3, n)) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil {
		return max(0, time.Until(at))
	}
	return 0
}
