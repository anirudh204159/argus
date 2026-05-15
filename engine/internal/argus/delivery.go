package argus

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	deliveryTimeout = 10 * time.Second
	userAgent       = "Argus/0.1"
)

var httpClient = &http.Client{
	Timeout: deliveryTimeout,
}

// deliveryResult captures the outcome of one delivery attempt.
type deliveryResult struct {
	Success      bool
	HTTPStatus   int // 0 if no response (e.g. connection error)
	LatencyMS    int64
	ErrorMessage string
}

// signPayload computes the HMAC-SHA256 signature of the payload using the subscription's secret.
// Returns hex-encoded signature suitable for an HTTP header.
func signPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// deliver performs ONE HTTP POST attempt to the subscription's webhook URL.
// Caller is responsible for retry logic.
func deliver(ctx context.Context, sub Subscription, ev Event, attempt int) deliveryResult {
	start := time.Now()
	result := deliveryResult{}

	// Build the JSON payload. We include enough info that consumers can dedupe
	// (event_id), verify (signature header), and act on the data (operation, before, after).
	payload, err := json.Marshal(map[string]any{
		"subscription_id": sub.ID,
		"table":           ev.Table,
		"schema":          ev.Schema,
		"operation":       ev.Operation,
		"before":          ev.Before,
		"after":           ev.After,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("marshal payload: %v", err)
		result.LatencyMS = time.Since(start).Milliseconds()
		return result
	}

	signature := signPayload(sub.HMACSecret, payload)

	req, err := http.NewRequestWithContext(ctx, "POST", sub.WebhookURL, bytes.NewReader(payload))
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("build request: %v", err)
		result.LatencyMS = time.Since(start).Milliseconds()
		return result
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Argus-Signature", "sha256="+signature)
	req.Header.Set("X-Argus-Subscription-Id", fmt.Sprintf("%d", sub.ID))
	req.Header.Set("X-Argus-Attempt", fmt.Sprintf("%d", attempt))

	resp, err := httpClient.Do(req)
	result.LatencyMS = time.Since(start).Milliseconds()

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("http: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.HTTPStatus = resp.StatusCode

	// Drain the body so the HTTP client can reuse the connection.
	// We don't care about the response body content for now.
	_, _ = io.Copy(io.Discard, resp.Body)

	// 2xx = success. Anything else = failure (caller retries).
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
	} else {
		result.ErrorMessage = fmt.Sprintf("non-2xx status: %d", resp.StatusCode)
	}

	return result
}
