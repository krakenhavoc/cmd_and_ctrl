package metrics

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// AddrEnv names the metrics listener's address. Unset is off. CD sets
// it to 127.0.0.1:9464 (ADR 0123 §10).
const AddrEnv = "CMDCTRL_METRICS_ADDR"

// ErrNotLoopback is the refusal of an address that is not loopback.
// The only reader is the monitoring agent on the same host, and a
// metrics page reachable from outside tells anyone the table counts.
var ErrNotLoopback = errors.New(AddrEnv + " must be a loopback address (127.0.0.0/8, ::1 or localhost) with a port")

// CheckAddr returns nil when addr is host:port and the host is a
// loopback IP or localhost. An empty host (":9464") listens on every
// interface and is refused, as is any other hostname, which cannot be
// checked without resolving it.
func CheckAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("%w: %q: %v", ErrNotLoopback, addr, err)
	}
	if port == "" {
		return fmt.Errorf("%w: %q has no port", ErrNotLoopback, addr)
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("%w: %q", ErrNotLoopback, addr)
}

// Listen checks addr, binds it, and checks the address it actually
// bound, which is what catches a "localhost" that resolves elsewhere.
// A refusal wraps ErrNotLoopback; any other error is the bind's.
func Listen(addr string) (net.Listener, error) {
	if err := CheckAddr(addr); err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if tcp, ok := ln.Addr().(*net.TCPAddr); !ok || !tcp.IP.IsLoopback() {
		_ = ln.Close()
		return nil, fmt.Errorf("%w: %q bound %s", ErrNotLoopback, addr, ln.Addr())
	}
	return ln, nil
}

// NewServer returns the metrics listener's own server: GET /metrics
// over g and nothing else. It is never mounted on the public mux, so no
// change to the public routes or to Caddy can expose it.
func NewServer(g prometheus.Gatherer, log *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{
		ErrorLog:      slogPrinter{log},
		ErrorHandling: promhttp.ContinueOnError,
		// One scraper, every 30 s. A pile-up means something is wrong,
		// and a gather takes room locks.
		MaxRequestsInFlight: 2,
		Timeout:             10 * time.Second,
	}))
	return &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// slogPrinter adapts slog to promhttp.Logger.
type slogPrinter struct{ log *slog.Logger }

func (p slogPrinter) Println(v ...any) {
	if p.log != nil {
		p.log.Error("metrics: serving /metrics", "err", fmt.Sprint(v...))
	}
}
