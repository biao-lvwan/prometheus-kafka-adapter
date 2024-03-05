package main

import "github.com/prometheus/client_golang/prometheus"

var (
	httpRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all http requests",
		})
	promBatches = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "incoming_prometheus_batches_total",
			Help: "Count of incoming prometheus batches (to be broken into individual metrics)",
		})
	serializeTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "serialized_total",
			Help: "Count of all serialization requests",
		})
	serializeFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "serialized_failed_total",
			Help: "Count of all serialization failures",
		})
	objectsFiltered = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "objects_filtered_total",
			Help: "Count of all filter attempts",
		})
	objectsWritten = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "objects_written_total",
			Help: "Count of all objects written to Kafka",
		})
	objectsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "objects_failed_total",
			Help: "Count of all objects write failures to Kafka",
		})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(promBatches)
	prometheus.MustRegister(serializeTotal)
	prometheus.MustRegister(serializeFailed)
	prometheus.MustRegister(objectsFiltered)
	prometheus.MustRegister(objectsFailed)
	prometheus.MustRegister(objectsWritten)
}
