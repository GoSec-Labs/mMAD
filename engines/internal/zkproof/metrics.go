package zkproof

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains all ZK proof engine metrics
type Metrics struct {
	// Generation metrics
	ProofRequests           *prometheus.CounterVec
	ProofGenerated          *prometheus.CounterVec
	ProofFailures           *prometheus.CounterVec
	ProofGenerationDuration *prometheus.HistogramVec

	// Verification metrics
	VerificationDuration prometheus.Histogram
	ValidProofs          prometheus.Counter
	InvalidProofs        prometheus.Counter
	VerificationFailures prometheus.Counter

	// System metrics
	ActiveProofs   prometheus.Gauge
	WorkerPoolSize prometheus.Gauge
	QueuedJobs     prometheus.Gauge

	// Circuit metrics
	CircuitCompilations *prometheus.CounterVec
	CircuitCacheHits    prometheus.Counter
	CircuitCacheMisses  prometheus.Counter
}

// NewMetrics creates new metrics
func NewMetrics() *Metrics {
	return &Metrics{
		ProofRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zkproof_requests_total",
				Help: "Total number of proof generation requests",
			},
			[]string{"type"},
		),

		ProofGenerated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zkproof_generated_total",
				Help: "Total number of proofs successfully generated",
			},
			[]string{"type"},
		),

		ProofFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zkproof_failures_total",
				Help: "Total number of proof generation failures",
			},
			[]string{"type"},
		),

		ProofGenerationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zkproof_generation_duration_seconds",
				Help:    "Time taken to generate proofs",
				Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s to ~17min
			},
			[]string{"type"},
		),

		VerificationDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "zkproof_verification_duration_seconds",
				Help:    "Time taken to verify proofs",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
			},
		),

		ValidProofs: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "zkproof_valid_total",
				Help: "Total number of valid proofs verified",
			},
		),

		InvalidProofs: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "zkproof_invalid_total",
				Help: "Total number of invalid proofs",
			},
		),

		VerificationFailures: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "zkproof_verification_failures_total",
				Help: "Total number of verification failures",
			},
		),

		ActiveProofs: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zkproof_active_generations",
				Help: "Number of proofs currently being generated",
			},
		),

		WorkerPoolSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zkproof_worker_pool_size",
				Help: "Size of the worker pool",
			},
		),
		QueuedJobs: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zkproof_queued_jobs",
				Help: "Number of jobs queued for processing",
			},
		),

		CircuitCompilations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zkproof_circuit_compilations_total",
				Help: "Total number of circuit compilations",
			},
			[]string{"circuit_id", "status"},
		),

		CircuitCacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "zkproof_circuit_cache_hits_total",
				Help: "Total number of circuit cache hits",
			},
		),

		CircuitCacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "zkproof_circuit_cache_misses_total",
				Help: "Total number of circuit cache misses",
			},
		),
	}
}
