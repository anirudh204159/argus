package argus

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// --- Engine metrics ---

var (
	binlogEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "argus_binlog_events_total",
			Help: "Total binlog events processed by the engine, by operation.",
		},
		[]string{"operation"}, // INSERT / UPDATE / DELETE
	)

	currentBinlogPos = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "argus_binlog_position_bytes",
			Help: "Current binlog position in the active file.",
		},
	)

	redisPublishesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "argus_redis_publishes_total",
			Help: "Total events published to Redis Stream.",
		},
	)

	redisPublishErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "argus_redis_publish_errors_total",
			Help: "Total Redis publish failures.",
		},
	)
)

// --- Worker metrics ---

var (
	deliveriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "argus_deliveries_total",
			Help: "Total delivery attempts, by outcome.",
		},
		[]string{"status"}, // success / failure
	)

	deliveryDurationMS = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "argus_delivery_duration_milliseconds",
			Help:    "Webhook delivery latency in milliseconds.",
			Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
	)

	retriesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "argus_retries_total",
			Help: "Total retry attempts.",
		},
	)

	dlqWritesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "argus_dlq_writes_total",
			Help: "Total events sent to dead-letter queue.",
		},
	)

	subscriptionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "argus_subscriptions_active",
			Help: "Current count of active subscriptions in the cache.",
		},
	)
)

// init registers all metrics with the default Prometheus registry on package load.
func init() {
	prometheus.MustRegister(
		binlogEventsTotal,
		currentBinlogPos,
		redisPublishesTotal,
		redisPublishErrorsTotal,
		deliveriesTotal,
		deliveryDurationMS,
		retriesTotal,
		dlqWritesTotal,
		subscriptionsActive,
	)
}

// serveMetrics starts an HTTP server on the given port that exposes /metrics, /healthz, /readyz.
// Runs as a goroutine.
func serveMetrics(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// For v0.1: same as healthz. Real readiness would check
		// MySQL + Redis connectivity. Easy follow-up.
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ready")
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Metrics server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("metrics server error: %v\n", err)
	}
}
