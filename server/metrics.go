package server

import "github.com/prometheus/client_golang/prometheus"

var (
	registerationsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "active_registrations",
		Namespace: "rendezvous",
		Help:      "Number of active unique registrations.",
	}, []string{"topic"})

	discoverySize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "discovery_size",
		Namespace: "rendezvous",
		Buckets:   []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Help:      "Number of records found for each discover requests.",
	}, []string{"topic"})

	discoveryDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "discovery_duration",
		Namespace: "rendezvous",
		Help:      "Discovery requests in seconds.",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10},
	}, []string{"topic"})

	errorsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "discovery_errors",
		Namespace: "rendezvous",
		Help:      "Number of errors labeled by the type of operation.",
	}, []string{"operation"})
)

func init() {
	prometheus.MustRegister(registerationsGauge, discoverySize, discoveryDuration, errorsCounter)
}
