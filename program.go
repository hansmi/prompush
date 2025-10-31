package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/cenkalti/backoff/v5"
	"github.com/prometheus/client_golang/prometheus/push"
)

type programOptions struct {
	URL string `env:"URL"`

	RootCAFile string `env:"CACERT_FILE"`
	CertFile   string `env:"CERT_FILE"`
	KeyFile    string `env:"KEY_FILE"`

	Retries    uint          `env:"RETRIES" envDefault:"2"`
	RetryDelay time.Duration `env:"RETRY_DELAY" envDefault:"10s"`

	Job         string `env:"JOB_NAME"`
	Instance    string `env:"INSTANCE_NAME"`
	MetricsFile string `env:"METRICS_FILE"`
}

func (o *programOptions) registerFlags(fs *flag.FlagSet, environ map[string]string) error {
	if err := env.ParseWithOptions(o, env.Options{
		Environment: environ,
		Prefix:      "PROMPUSH_",
	}); err != nil {
		return err
	}

	fs.StringVar(&o.URL, "gateway", o.URL,
		"Pushgateway URL (e.g. https://pushgateway.example.com:9091). Defaults to $PROMPUSH_URL.")

	fs.StringVar(&o.RootCAFile, "cacert", o.RootCAFile,
		"Path to CA certificate file for server verification. Defaults to $PROMPUSH_CACERT_FILE.")
	fs.StringVar(&o.CertFile, "cert", o.CertFile,
		"Path to client certificate file. Defaults to $PROMPUSH_CERT_FILE.")
	fs.StringVar(&o.KeyFile, "key", o.KeyFile,
		"Path to client private key file. Defaults to $PROMPUSH_KEY_FILE.")

	fs.UintVar(&o.Retries, "retries", o.Retries,
		"Number of retries for transient failures. Defaults to $PROMPUSH_RETRIES.")
	fs.DurationVar(&o.RetryDelay, "retry_delay", o.RetryDelay,
		"Initial delay between push retries. Defaults to $PROMPUSH_RETRY_DELAY.")

	fs.StringVar(&o.Job, "job", o.Job,
		"Job label for the metrics. Defaults to $PROMPUSH_JOB_NAME.")
	fs.StringVar(&o.Instance, "instance", o.Instance,
		"Instance label (e.g. server01). Defaults to $PROMPUSH_INSTANCE_NAME.")
	fs.StringVar(&o.MetricsFile, "metrics", o.MetricsFile,
		"Path to the file containing metrics in Prometheus text exposition format. Defaults to $PROMPUSH_METRICS_FILE.")

	return nil
}

type program struct {
	retries    uint
	retryDelay time.Duration

	pusher *push.Pusher
}

func newProgram(opts programOptions) (*program, error) {
	if opts.URL == "" {
		return nil, errors.New("gateway URL is required")
	}

	if opts.Job == "" {
		return nil, errors.New("job name is required")
	}

	if opts.MetricsFile == "" {
		return nil, errors.New("metrics file is required")
	}

	parsedUrl, err := url.Parse(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("gateway URL: %w", err)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if opts.RootCAFile != "" {
		content, err := os.ReadFile(opts.RootCAFile)
		if err != nil {
			return nil, fmt.Errorf("CA certificate file: %w", err)
		}

		pool := x509.NewCertPool()

		if !pool.AppendCertsFromPEM(content) {
			log.Printf("WARNING: Failed to parse CA certificates read from %q.", opts.RootCAFile)
		}

		tlsConfig.RootCAs = pool
	}

	if opts.CertFile != "" || opts.KeyFile != "" {
		clientCert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading client key %q, certificate %q: %w", opts.CertFile, opts.KeyFile, err)
		}

		tlsConfig.Certificates = append(tlsConfig.Certificates, clientCert)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	metrics, err := readMetrics(opts.MetricsFile)
	if err != nil {
		return nil, fmt.Errorf("metrics from %q: %w", opts.MetricsFile, err)
	}

	pusher := push.New(parsedUrl.String(), opts.Job).
		Client(httpClient).
		Gatherer(gatherer(metrics))

	if opts.Instance != "" {
		pusher.Grouping("instance", opts.Instance)
	}

	return &program{
		retries:    opts.Retries,
		retryDelay: opts.RetryDelay,
		pusher:     pusher,
	}, nil
}

func isRetryable(err error) bool {
	return true
}

func (p *program) run(ctx context.Context) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = p.retryDelay
	b.MaxInterval = 10 * p.retryDelay
	b.RandomizationFactor = 0.1

	op := func() (struct{}, error) {
		err := p.pusher.PushContext(ctx)

		if err != nil && !isRetryable(err) {
			err = backoff.Permanent(err)
		}

		return struct{}{}, err
	}

	notify := func(err error, delay time.Duration) {
		log.Printf("Retrying failed push in %s: %v", delay.Truncate(100*time.Millisecond), err)
	}

	_, err := backoff.Retry(ctx, op,
		backoff.WithBackOff(b),
		backoff.WithNotify(notify),
		backoff.WithMaxTries(1+p.retries),
	)

	return err
}
