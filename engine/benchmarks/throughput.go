package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	sourceDSN       = "root:rootpass@tcp(127.0.0.1:3307)/argus_demo"
	numEvents       = 100
	receiverAddr    = ":8080"
	receiverPath    = "/webhook"
	receiverTimeout = 60 * time.Second
)

type record struct {
	id         int
	insertedAt time.Time
	receivedAt time.Time
}

var (
	recordsMu sync.Mutex
	records   = make(map[int]*record) // id -> record

	receivedChan = make(chan int, numEvents+10)
)

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP receiver
	go startReceiver()
	time.Sleep(200 * time.Millisecond) // let server bind

	// Connect to source MySQL
	db, err := sql.Open("mysql", sourceDSN)
	if err != nil {
		fmt.Println("db open:", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Println("db ping:", err)
		return
	}

	// Clean up any old data
	_, _ = db.Exec("DELETE FROM orders WHERE id >= 10000 AND id < 20000")

	fmt.Printf("Inserting %d rows...\n", numEvents)

	// Insert N rows, recording insert timestamps
	for i := 0; i < numEvents; i++ {
		id := 10000 + i
		now := time.Now()

		recordsMu.Lock()
		records[id] = &record{id: id, insertedAt: now}
		recordsMu.Unlock()

		_, err := db.Exec("INSERT INTO orders VALUES (?, ?, ?)", id, 1.00, fmt.Sprintf("bench-%d", i))
		if err != nil {
			fmt.Printf("insert %d error: %v\n", id, err)
			continue
		}

		// Small spacing so we don't slam MySQL
		time.Sleep(20 * time.Millisecond)
	}

	fmt.Printf("Inserts done. Waiting for webhooks (timeout %v)...\n", receiverTimeout)

	// Wait for all webhooks or timeout
	timeout := time.After(receiverTimeout)
	received := 0

waitLoop:
	for received < numEvents {
		select {
		case <-receivedChan:
			received++
			if received%10 == 0 {
				fmt.Printf("  received %d/%d\n", received, numEvents)
			}
		case <-timeout:
			fmt.Printf("Timeout — received only %d/%d\n", received, numEvents)
			break waitLoop
		}
	}

	cancel()
	printResults(received)
}

func startReceiver() {
	mux := http.NewServeMux()
	mux.HandleFunc(receiverPath, func(w http.ResponseWriter, r *http.Request) {
		receivedAt := time.Now()
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		// Parse the ID out of the payload — quick string search avoids JSON dependency
		id := parseID(string(body))
		if id < 10000 {
			w.WriteHeader(http.StatusOK)
			return
		}

		recordsMu.Lock()
		if rec, ok := records[id]; ok && rec.receivedAt.IsZero() {
			rec.receivedAt = receivedAt
		}
		recordsMu.Unlock()

		receivedChan <- id
		w.WriteHeader(http.StatusOK)
	})

	fmt.Printf("Receiver listening on %s%s\n", receiverAddr, receiverPath)
	if err := http.ListenAndServe(receiverAddr, mux); err != nil {
		fmt.Printf("receiver error: %v\n", err)
	}
}

// parseID extracts the "id" field from the JSON payload.
// We could use encoding/json but this is faster + dependency-free for benchmarks.
func parseID(body string) int {
	const marker = `"id":`
	start := -1
	for i := 0; i+len(marker) < len(body); i++ {
		if body[i:i+len(marker)] == marker {
			start = i + len(marker)
			break
		}
	}
	if start == -1 {
		return 0
	}
	// skip whitespace
	for start < len(body) && (body[start] == ' ' || body[start] == '\t') {
		start++
	}
	end := start
	for end < len(body) && body[end] >= '0' && body[end] <= '9' {
		end++
	}
	if start == end {
		return 0
	}
	id := 0
	for _, c := range body[start:end] {
		id = id*10 + int(c-'0')
	}
	return id
}

func printResults(received int) {
	recordsMu.Lock()
	defer recordsMu.Unlock()

	latencies := []int64{} // milliseconds
	for _, rec := range records {
		if rec.receivedAt.IsZero() {
			continue
		}
		latencies = append(latencies, rec.receivedAt.Sub(rec.insertedAt).Milliseconds())
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Printf("Events inserted:     %d\n", numEvents)
	fmt.Printf("Events received:     %d\n", received)
	fmt.Printf("Loss rate:           %.2f%%\n", float64(numEvents-received)/float64(numEvents)*100)
	if len(latencies) == 0 {
		fmt.Println("No latency data.")
		return
	}
	fmt.Printf("p50 latency:         %d ms\n", percentile(latencies, 50))
	fmt.Printf("p95 latency:         %d ms\n", percentile(latencies, 95))
	fmt.Printf("p99 latency:         %d ms\n", percentile(latencies, 99))
	fmt.Printf("Max latency:         %d ms\n", latencies[len(latencies)-1])
	fmt.Printf("Mean latency:        %d ms\n", mean(latencies))
}

func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func mean(xs []int64) int64 {
	if len(xs) == 0 {
		return 0
	}
	var total int64
	for _, x := range xs {
		total += x
	}
	return total / int64(len(xs))
}
