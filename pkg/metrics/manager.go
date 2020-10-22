package metrics

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dmitsh/pushgatewayredis/pkg/config"
	"github.com/dmitsh/pushgatewayredis/pkg/redis"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsCacheSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "metrics_cache_size",
			Help: "Number of metrics in cache.",
		})

	metricsRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "metrics_request_duration_seconds",
			Help:    "Time (in seconds) spent serving metrics requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status_code"})
)

type MetricsManager struct {
	server *http.Server
	cfg    *config.Config
	db     *redis.RedisClient
	logger log.Logger
}

func NewMetricsManager(logger log.Logger, cfg *config.Config, db *redis.RedisClient) *MetricsManager {
	mm := &MetricsManager{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.MetricsPath, mm.serveMetricsPath)
	mux.HandleFunc(cfg.IngestPath, mm.serveIngestPath)
	mux.Handle(cfg.TelemetryPath, promhttp.Handler())
	mux.HandleFunc("/", mm.serveDefault)

	mm.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}
	return mm
}

func (mm *MetricsManager) Run() error {
	if mm.cfg.TLSEnabled {
		return mm.server.ListenAndServeTLS(mm.cfg.TLSCertPath, mm.cfg.TLSKeyPath)
	}
	return mm.server.ListenAndServe()
}

func (mm *MetricsManager) Close(ctx context.Context) error {
	return mm.server.Shutdown(ctx)
}

func (mm *MetricsManager) serveMetricsPath(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.Method != "GET" {
		metricsRequestDuration.WithLabelValues("405").Observe(time.Now().Sub(start).Seconds())
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	metrics, err := mm.db.GetAll(r.Context())
	if err != nil {
		level.Error(mm.logger).Log("msg", "Redis error", "err", err)
		metricsRequestDuration.WithLabelValues("500").Observe(time.Now().Sub(start).Seconds())
		http.Error(w, "Redis error", http.StatusInternalServerError)
		return
	}
	metricsRequestDuration.WithLabelValues("200").Observe(time.Now().Sub(start).Seconds())
	metricsCacheSize.Set(float64(len(metrics)))
	fmt.Fprintf(w, strings.Join(metrics, "\n"))
}

func (mm *MetricsManager) serveIngestPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	scanner := bufio.NewScanner(r.Body)
	keys, vals := []string{}, []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		if indx := strings.LastIndex(line, " "); indx != -1 {
			keys = append(keys, line[:indx])
			vals = append(vals, line[indx:])
		}
	}
	if err := scanner.Err(); err != nil {
		level.Error(mm.logger).Log("msg", "Read error", "err", err)
		http.Error(w, "Cannot read body", http.StatusBadRequest)
		return
	}
	if err := mm.db.MSet(r.Context(), keys, vals); err != nil {
		level.Error(mm.logger).Log("msg", "Redis error", "err", err)
		http.Error(w, "Redis error", http.StatusInternalServerError)
		return
	}
}

func (mm *MetricsManager) serveDefault(w http.ResponseWriter, r *http.Request) {
	level.Error(mm.logger).Log("msg", "Unsupported path", "url", r.URL.Path)
	http.Error(w, "404 page not found.", http.StatusNotFound)
}
