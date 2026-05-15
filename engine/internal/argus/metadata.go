package argus

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	metadataDSN  = "root:rootpass@tcp(127.0.0.1:3308)/argus_metadata"
	refreshEvery = 30 * time.Second
)

// Subscription represents a row from the subscriptions table.
type Subscription struct {
	ID         int64
	UserID     int64
	SourceID   int64
	Name       string
	Tables     []string
	Operations []string
	WebhookURL string
	HMACSecret string
	RetryMax   int
	Active     bool
}

var (
	metadataDB *sql.DB

	subscriptionsMu     sync.RWMutex
	cachedSubscriptions []Subscription
)

func initMetadataDB() error {
	var err error
	metadataDB, err = sql.Open("mysql", metadataDSN)
	if err != nil {
		return fmt.Errorf("metadata db open: %w", err)
	}
	if err := metadataDB.Ping(); err != nil {
		return fmt.Errorf("metadata db ping: %w", err)
	}
	return nil
}

// loadSubscriptions fetches active subscriptions from the metadata DB.
func loadSubscriptions(ctx context.Context) ([]Subscription, error) {
	rows, err := metadataDB.QueryContext(ctx, `
		SELECT id, user_id, source_id, name, tables, operations,
		       webhook_url, hmac_secret, retry_max, active
		FROM subscriptions
		WHERE active = TRUE
	`)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		var tablesJSON, opsJSON []byte

		if err := rows.Scan(
			&s.ID, &s.UserID, &s.SourceID, &s.Name,
			&tablesJSON, &opsJSON,
			&s.WebhookURL, &s.HMACSecret, &s.RetryMax, &s.Active,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		if err := json.Unmarshal(tablesJSON, &s.Tables); err != nil {
			return nil, fmt.Errorf("parse tables for sub %d: %w", s.ID, err)
		}
		if err := json.Unmarshal(opsJSON, &s.Operations); err != nil {
			return nil, fmt.Errorf("parse operations for sub %d: %w", s.ID, err)
		}

		subs = append(subs, s)
	}

	return subs, rows.Err()
}

// refreshSubscriptions periodically reloads the subscription cache.
// Runs as a goroutine for the lifetime of the worker.
func refreshSubscriptions(ctx context.Context) {
	ticker := time.NewTicker(refreshEvery)
	defer ticker.Stop()

	for {
		subs, err := loadSubscriptions(ctx)
		if err != nil {
			fmt.Println("subscription refresh error:", err)
		} else {
			subscriptionsMu.Lock()
			cachedSubscriptions = subs
			subscriptionsMu.Unlock()
			fmt.Printf("loaded %d active subscriptions\n", len(subs))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// getSubscriptions returns a snapshot of the current subscription cache.
func getSubscriptions() []Subscription {
	subscriptionsMu.RLock()
	defer subscriptionsMu.RUnlock()
	// Return a copy so callers don't accidentally mutate the cache
	snapshot := make([]Subscription, len(cachedSubscriptions))
	copy(snapshot, cachedSubscriptions)
	return snapshot
}

// logDelivery records one delivery attempt in the delivery_log table.
// logDelivery records one delivery attempt in the delivery_log table.
func logDelivery(ctx context.Context, subscriptionID int64, eventID string, attempt int, result deliveryResult) {
	status := "failure"
	if result.Success {
		status = "success"
	}

	var httpStatus sql.NullInt32
	if result.HTTPStatus != 0 {
		httpStatus = sql.NullInt32{Int32: int32(result.HTTPStatus), Valid: true}
	}

	var errorMsg sql.NullString
	if result.ErrorMessage != "" {
		errorMsg = sql.NullString{String: result.ErrorMessage, Valid: true}
	}

	_, err := metadataDB.ExecContext(ctx, `
		INSERT INTO delivery_log
			(subscription_id, event_id, attempt, status, http_status, latency_ms, error_message, delivered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`,
		subscriptionID, eventID, attempt, status, httpStatus, int(result.LatencyMS), errorMsg,
	)
	if err != nil {
		fmt.Printf("delivery log write error: %v\n", err)
	}
}
