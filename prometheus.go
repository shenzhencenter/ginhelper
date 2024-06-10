package ginhelper

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shenzhencenter/goset"
)

var (
	skipMetricsPaths    *goset.Set[string]
	counter             *prometheus.CounterVec
	histogram           *prometheus.HistogramVec
	ginMetricsNamespace = os.Getenv("GIN_METRICS_NAMESPACE")
)

func counterVec() *prometheus.CounterVec {
	if counter != nil {
		return counter
	}
	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ginMetricsNamespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests made.",
		},
		[]string{"status", "host", "method", "path"},
	)
	prometheus.MustRegister(counter)
	return counter
}

func histogramVec() *prometheus.HistogramVec {
	if histogram != nil {
		return histogram
	}
	histogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: ginMetricsNamespace,
			Name:      "http_request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
		},
		[]string{"status", "host", "method", "path"},
	)
	prometheus.MustRegister(histogram)
	return histogram
}

func inc(httpStatus, host, method, path string) {
	counterVec().WithLabelValues(httpStatus, host, method, path).Inc()
}

func observe(handlsTime float64, httpStatus, host, method, path string) {
	histogramVec().WithLabelValues(httpStatus, host, method, path).Observe(handlsTime)
}

func GinMetricsMiddleware(skip ...string) gin.HandlerFunc {
	skipMetricsPaths = goset.New(skip...)
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		handlsTime := time.Since(start).Seconds()
		path := c.FullPath()
		if len(path) == 0 && c.Request.URL != nil {
			path = c.Request.URL.Path
		}
		if hit, _ := skipMetricsPaths.SearchOne(func(prefix string) bool { return strings.HasPrefix(path, prefix) }); hit {
			return
		}
		httpStatus := strconv.Itoa(c.Writer.Status())
		host := c.Request.Host
		if len(host) == 0 && c.Request.URL != nil {
			host = c.Request.URL.Host
		}
		method := c.Request.Method
		inc(httpStatus, host, method, path)
		observe(handlsTime, httpStatus, host, method, path)
	}
}

func GinMetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
