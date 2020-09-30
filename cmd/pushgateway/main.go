package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/dmitsh/pushgatewayredis/pkg/config"
	"github.com/dmitsh/pushgatewayredis/pkg/metrics"
	"github.com/dmitsh/pushgatewayredis/pkg/redis"
)

func main() {
	var (
		configFile string
		cfg        config.Config
	)

	a := kingpin.New(filepath.Base(os.Args[0]), "The Prometheus Redis Pushgateway")
	a.HelpFlag.Short('h')
	a.Flag("config.file", "Prometheus configuration file path.").StringVar(&configFile)
	a.Flag("port", "Service port.").Short('p').Default("9753").IntVar(&cfg.Port)
	a.Flag("metrics.path", "Metrics path.").Short('m').Default("/metrics").StringVar(&cfg.MetricsPath)
	a.Flag("ingest.path", "Ingest path.").Short('i').Default("/ingest").StringVar(&cfg.IngestPath)
	a.Flag("redis.endpoint", "Redis endpoint(s).").Default(":6379").StringVar(&cfg.RedisConfig.Endpoint)
	a.Flag("redis.expiration", "Redis key/value expiration.").Default("5m").DurationVar(&cfg.RedisConfig.Expiration)
	a.Flag("tls.enabled", "Enable TLS.").Default("false").BoolVar(&cfg.TLSEnabled)
	a.Flag("tls.key", "Path to the server key.").StringVar(&cfg.TLSKeyPath)
	a.Flag("tls.cert", "Path to the server certificate.").StringVar(&cfg.TLSCertPath)

	logConfig := &promlog.Config{}
	promlogflag.AddFlags(a, logConfig)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	ctx := context.Background()
	logger := promlog.New(logConfig)

	if len(configFile) > 0 {
		if err := cfg.LoadFile(configFile); err != nil {
			level.Error(logger).Log("msg", "Error loading config", "path", configFile, "err", err)
			os.Exit(1)
		}
	}

	db := redis.NewRedisClient(&cfg.RedisConfig)
	mm := metrics.NewMetricsManager(logger, &cfg, db)

	var g run.Group
	{
		// Termination handler
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancel := make(chan struct{})
		g.Add(
			func() error {
				select {
				case <-term:
					level.Warn(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
				case <-cancel:
				}
				return nil
			},
			func(err error) {
				close(cancel)
			},
		)
	}
	{
		// Redis
		cancel := make(chan struct{})
		g.Add(
			func() error {
				level.Info(logger).Log("msg", "Starting Redis client...")

				if err := db.Ping(ctx); err != nil {
					level.Error(logger).Log("mgs", "Failed to start Redis", "err", err)
					return err
				}
				<-cancel
				return nil
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping Redis client...")
				db.Close()
				close(cancel)
			},
		)
	}
	{
		// Metrics manager
		g.Add(
			func() error {
				level.Info(logger).Log("msg", "Starting Metrics manager...")
				return mm.Run()
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping Metrics manager...")
				mm.Close(ctx)
			},
		)
	}
	if err := g.Run(); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
