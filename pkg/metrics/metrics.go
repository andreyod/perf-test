package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TimePoints = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "watch_latency_time_point",
		Help: "Time points for request sending/returned and for event recieved",
	}, []string{"stage"})
)

func RegisterWatchMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(TimePoints)
	panic(http.ListenAndServe(":8080", nil))
}

