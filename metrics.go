package main

import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Count of all http requests",
        })
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
}
