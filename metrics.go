package main

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	targetUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "healthmonitor_target_up",
		Help: "1 if the last check of the target succeeded, 0 otherwise.",
	}, []string{"target", "url"})

	responseSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "healthmonitor_response_seconds",
		Help:    "Response time of the target in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"target"})

	checksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "healthmonitor_checks_total",
		Help: "Total number of checks, split by result.",
	}, []string{"target", "result"})
)

// record updates all Prometheus metrics for one finished check and logs it.
func record(t Target, r Result) {
	result := "failure"
	up := 0.0
	if r.Up {
		result = "success"
		up = 1.0
	}

	targetUp.WithLabelValues(t.Name, t.URL).Set(up)
	checksTotal.WithLabelValues(t.Name, result).Inc()
	if r.Err == nil {
		responseSeconds.WithLabelValues(t.Name).Observe(r.Duration.Seconds())
	}

	slog.Info("check finished",
		"target", t.Name,
		"up", r.Up,
		"status", r.StatusCode,
		"duration_ms", r.Duration.Milliseconds(),
		"error", r.Err,
	)
}
