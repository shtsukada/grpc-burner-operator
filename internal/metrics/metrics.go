package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpcburner_operator_reconcile_total",
			Help: "Total number of reconciliations performed",
		},
		[]string{"controller"},
	)

	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpcburner_operator_reconcile_errors_total",
			Help: "Total number of reconciliation errors",
		},
		[]string{"controller"},
	)
)

func RegisterCustomMetrics() {
	prometheus.MustRegister(ReconcileTotal, ReconcileErrors)
}
