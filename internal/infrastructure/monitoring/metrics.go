package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	AnalysisJobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analysis_jobs_total",
			Help: "Total number of analysis jobs",
		},
		[]string{"status", "url_type"},
	)

	AnalysisJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analysis_job_duration_seconds",
			Help:    "Analysis job duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"status"},
	)

	QueueLength = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_length",
			Help: "Current queue length",
		},
		[]string{"queue_name"},
	)

	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
		},
		[]string{"name"},
	)

	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	RedisConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_connections_active",
			Help: "Number of active Redis connections",
		},
	)
)

func RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	status := prometheus.Labels{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": string(rune(statusCode)),
	}

	HTTPRequestDuration.With(status).Observe(duration.Seconds())
	HTTPRequestsTotal.With(status).Inc()
}

func RecordAnalysisJob(status, urlType string, duration time.Duration) {
	AnalysisJobsTotal.With(prometheus.Labels{
		"status":   status,
		"url_type": urlType,
	}).Inc()

	AnalysisJobDuration.With(prometheus.Labels{
		"status": status,
	}).Observe(duration.Seconds())
}

func UpdateQueueLength(queueName string, length int64) {
	QueueLength.With(prometheus.Labels{
		"queue_name": queueName,
	}).Set(float64(length))
}

func RecordCacheHit(cacheType string) {
	CacheHitsTotal.With(prometheus.Labels{
		"cache_type": cacheType,
	}).Inc()
}

func RecordCacheMiss(cacheType string) {
	CacheMissesTotal.With(prometheus.Labels{
		"cache_type": cacheType,
	}).Inc()
}

func UpdateCircuitBreakerState(name string, state int) {
	CircuitBreakerState.With(prometheus.Labels{
		"name": name,
	}).Set(float64(state))
}

func SetDatabaseConnections(active int) {
	DatabaseConnectionsActive.Set(float64(active))
}

func SetRedisConnections(active int) {
	RedisConnectionsActive.Set(float64(active))
}
