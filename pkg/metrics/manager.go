package metrics

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dmitsh/pushgatewayredis/pkg/config"
	"github.com/dmitsh/pushgatewayredis/pkg/redis"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
	mm.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mm,
	}
	return mm
}

func (mm *MetricsManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case mm.cfg.MetricsPath:
		if r.Method != "GET" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}
		metrics, err := mm.db.GetAll(r.Context())
		if err != nil {
			level.Error(mm.logger).Log("msg", "Redis error", "err", err)
			http.Error(w, "redis error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, strings.Join(metrics, "\n"))
	case mm.cfg.IngestPath:
		if r.Method != "POST" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
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
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		if err := mm.db.MSet(r.Context(), keys, vals); err != nil {
			level.Error(mm.logger).Log("msg", "Redis error", "err", err)
			http.Error(w, "redis error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "404 not found.", http.StatusNotFound)
	}
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
