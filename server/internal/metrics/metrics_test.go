package metrics

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// nested builds the server's shape in miniature: an outer mux with a
// health route, a prefix mounted behind a middleware that copies the
// request (as auth.Middleware and the rate limiter do), and an inner
// mux mounted at "/" (as the lobby is).
func nested() http.Handler {
	cards := NewServeMux()
	cards.HandleFunc("GET /t/cards/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	copyCtx := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), copyKey{}, 1)))
		})
	}

	lobby := NewServeMux()
	lobby.Handle("POST /t/games/{id}/join", copyCtx(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})))
	lobby.HandleFunc("GET /t/me", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{}")) })

	outer := NewServeMux()
	outer.HandleFunc("GET /t/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	outer.Handle("/t/cards/", copyCtx(cards))
	outer.Handle("/", lobby)
	return InstrumentHTTP(outer, "GET /t/healthz")
}

// copyKey is the test middleware's own context key.
type copyKey struct{}

func count(route, method, code string) float64 {
	return testutil.ToFloat64(httpRequests.WithLabelValues(route, method, code))
}

func serve(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

// The route holder labels each request with the INNERMOST mux's
// pattern: a lobby route with its own pattern, not the "/" it is
// mounted on; a card route through a request-copying middleware with
// the card mux's pattern; and a path the lobby does not know as
// unmatched, not "/".
func TestRouteIsTheInnermostMatchedPattern(t *testing.T) {
	h := nested()
	cases := []struct {
		method, path, route, code string
	}{
		{"POST", "/t/games/3f2a/join", "POST /t/games/{id}/join", "4xx"},
		{"GET", "/t/me", "GET /t/me", "2xx"},
		{"GET", "/t/healthz", "GET /t/healthz", "2xx"},
		{"GET", "/t/cards/abc", "GET /t/cards/{id}", "4xx"},
		// The outer mux matched "/t/cards/", the card mux nothing.
		{"GET", "/t/cards/abc/def/ghi", RouteUnmatched, "4xx"},
		// The outer mux matched "/", the lobby mux nothing.
		{"GET", "/t/no-such-route", RouteUnmatched, "4xx"},
		// 405 from the lobby mux: no pattern.
		{"DELETE", "/t/me", RouteUnmatched, "4xx"},
	}
	for _, c := range cases {
		before := count(c.route, c.method, c.code)
		serve(h, c.method, c.path)
		if got := count(c.route, c.method, c.code) - before; got != 1 {
			t.Errorf("%s %s: cmdctrl_http_requests_total{route=%q,method=%q,code=%q} rose by %v, want 1",
				c.method, c.path, c.route, c.method, c.code, got)
		}
	}
	before := count("/", "GET", "4xx")
	serve(h, "GET", "/t/nope")
	if count("/", "GET", "4xx") != before {
		t.Error(`a lobby 404 was labelled route="/"`)
	}
}

// A route passed as untimed is counted but not observed.
func TestUntimedRouteIsNotObserved(t *testing.T) {
	h := nested()
	reg := prometheus.NewRegistry()
	reg.MustRegister(httpSeconds)
	timed := func(route string) uint64 {
		families, err := reg.Gather()
		if err != nil {
			t.Fatal(err)
		}
		var n uint64
		for _, f := range families {
			for _, m := range f.GetMetric() {
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "route" && lp.GetValue() == route {
						n += m.GetHistogram().GetSampleCount()
					}
				}
			}
		}
		return n
	}
	hb, mb := timed("GET /t/healthz"), timed("GET /t/me")
	serve(h, "GET", "/t/healthz")
	serve(h, "GET", "/t/me")
	if got := timed("GET /t/healthz") - hb; got != 0 {
		t.Errorf("untimed route observed %d times", got)
	}
	if got := timed("GET /t/me") - mb; got != 1 {
		t.Errorf("timed route observed %d times, want 1", got)
	}
}

func TestCodeClass(t *testing.T) {
	for status, want := range map[int]string{
		0: "2xx", 200: "2xx", 204: "2xx", 299: "2xx",
		301: "3xx", 304: "3xx", 307: "3xx",
		400: "4xx", 401: "4xx", 404: "4xx", 429: "4xx",
		500: "5xx", 503: "5xx", 599: "5xx",
	} {
		if got := codeClass(status); got != want {
			t.Errorf("codeClass(%d) = %q, want %q", status, got, want)
		}
	}
}

// The status class as a handler produces it, end to end: the first
// final status wins, an informational one does not count, writing
// without a status is 200, a hijack (a WebSocket upgrade) is 2xx, and
// a panic is 5xx.
func TestCodeClassFromHandlers(t *testing.T) {
	cases := []struct {
		name string
		h    http.HandlerFunc
		want string
	}{
		{"nothing written", func(http.ResponseWriter, *http.Request) {}, "2xx"},
		{"body only", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("x")) }, "2xx"},
		{"redirect", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) }, "3xx"},
		{"early hints then 503", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusEarlyHints)
			w.WriteHeader(http.StatusServiceUnavailable)
		}, "5xx"},
		{"404 then a second WriteHeader", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.WriteHeader(http.StatusOK)
		}, "4xx"},
		{"flush", func(w http.ResponseWriter, _ *http.Request) { w.(http.Flusher).Flush() }, "2xx"},
	}
	for _, c := range cases {
		mux := NewServeMux()
		pattern := "GET /t/code/" + strings.ReplaceAll(c.name, " ", "-")
		mux.Handle(pattern, c.h)
		before := count(pattern, "GET", c.want)
		serve(InstrumentHTTP(mux), "GET", strings.TrimPrefix(pattern, "GET "))
		if got := count(pattern, "GET", c.want) - before; got != 1 {
			t.Errorf("%s: %s count rose by %v, want 1", c.name, c.want, got)
		}
	}

	t.Run("panic", func(t *testing.T) {
		mux := NewServeMux()
		mux.HandleFunc("GET /t/code/panic", func(http.ResponseWriter, *http.Request) { panic("boom") })
		before := count("GET /t/code/panic", "GET", "5xx")
		func() {
			defer func() {
				if recover() == nil {
					t.Error("the panic was swallowed; it must reach net/http")
				}
			}()
			serve(InstrumentHTTP(mux), "GET", "/t/code/panic")
		}()
		if got := count("GET /t/code/panic", "GET", "5xx") - before; got != 1 {
			t.Errorf("panic: 5xx count rose by %v, want 1", got)
		}
	})

	t.Run("hijack", func(t *testing.T) {
		mux := NewServeMux()
		mux.HandleFunc("GET /t/code/hijack", func(w http.ResponseWriter, _ *http.Request) {
			conn, brw, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("hijack through the status writer: %v", err)
				return
			}
			_, _ = brw.WriteString("HTTP/1.1 101 Switching Protocols\r\n\r\n")
			_ = brw.Flush()
			_ = conn.Close()
		})
		// The count lands after the handler returns, which is after the
		// client has its answer: wait for the instrumented handler.
		done := make(chan struct{})
		inst := InstrumentHTTP(mux)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inst.ServeHTTP(w, r)
			close(done)
		}))
		defer srv.Close()
		before := count("GET /t/code/hijack", "GET", "2xx")
		resp, err := http.Get(srv.URL + "/t/code/hijack")
		if err == nil {
			_ = resp.Body.Close()
		}
		<-done
		if got := count("GET /t/code/hijack", "GET", "2xx") - before; got != 1 {
			t.Errorf("hijack: 2xx count rose by %v, want 1", got)
		}
	})
}

// A method outside the standard set is "other": the method is text the
// client chose.
func TestMethodIsAClosedSet(t *testing.T) {
	h := nested()
	before := count(RouteUnmatched, "other", "4xx")
	serve(h, "BREW", "/t/me")
	if got := count(RouteUnmatched, "other", "4xx") - before; got != 1 {
		t.Errorf("method BREW: other count rose by %v, want 1", got)
	}
}

// A pattern the mux did not register never becomes a label, whatever
// r.Pattern says: CONNECT redirects report a path there.
func TestUnregisteredPatternIsUnmatched(t *testing.T) {
	m := NewServeMux()
	m.HandleFunc("GET /t/own", func(http.ResponseWriter, *http.Request) {})
	if got := m.own("/t/some/client/path/"); got != "" {
		t.Errorf("own(unregistered) = %q, want empty", got)
	}
	if got := m.own("GET /t/own"); got != "GET /t/own" {
		t.Errorf("own(registered) = %q", got)
	}
}

// TestMetricLabelsAreClosedSets (ADR 0123 §3, §11) gathers the whole
// registry, after driving every route shape above so the HTTP series
// exist, and fails on a label name outside the table, a value outside
// its set, or a banned substring in a metric or label name.
func TestMetricLabelsAreClosedSets(t *testing.T) {
	h := nested()
	for _, p := range []struct{ m, path string }{
		{"POST", "/t/games/x/join"}, {"GET", "/t/me"}, {"GET", "/t/healthz"},
		{"GET", "/t/cards/x"}, {"GET", "/t/zzz"}, {"BREW", "/t/me"},
	} {
		serve(h, p.m, p.path)
	}
	if err := CheckClosedLabels(Registry); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
	// The upstream collectors are really there, so the guard read them.
	families, err := Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, f := range families {
		seen[f.GetName()] = true
	}
	for _, want := range []string{"cmdctrl_build_info", "cmdctrl_http_requests_total", "cmdctrl_http_request_seconds", "go_goroutines", "go_info", "process_start_time_seconds"} {
		if !seen[want] {
			t.Errorf("registry has no %s", want)
		}
	}
}

// The guard catches each way of breaking the rule.
func TestClosedLabelGuardCatchesViolations(t *testing.T) {
	cases := []struct {
		name string
		c    prometheus.Collector
		want string
	}{
		{"label outside the table", gaugeWith("cmdctrl_t_one", "colour", "red"), `label "colour" is not in the label table`},
		{"value outside its set", gaugeWith("cmdctrl_t_two", "code", "418"), `code="418" is outside its set`},
		{"route that is a path", gaugeWith("cmdctrl_t_three", "route", "/games/3f2a/join"), `route="/games/3f2a/join" is outside its set`},
		{"banned metric name", prometheus.NewGauge(prometheus.GaugeOpts{Name: "cmdctrl_t_by_remote"}), `metric name contains "remote"`},
		{"banned label name", gaugeWith("cmdctrl_t_four", "game_id", "x"), `label name "game_id" contains "id"`},
		{"foreign family", prometheus.NewGauge(prometheus.GaugeOpts{Name: "promhttp_t_things"}), "not a cmdctrl_ metric"},
	}
	for _, c := range cases {
		reg := prometheus.NewRegistry()
		reg.MustRegister(c.c)
		err := CheckClosedLabels(reg)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v, want an error containing %q", c.name, err, c.want)
		}
	}
}

func gaugeWith(name, label, value string) prometheus.Collector {
	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: "test"}, []string{label})
	g.WithLabelValues(value).Set(1)
	return g
}

func TestBuildInfo(t *testing.T) {
	if BuildCommit() == "" {
		t.Fatal("BuildCommit is empty; it must be a revision or unknown")
	}
	families, err := NewRegistry().Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range families {
		if f.GetName() != "cmdctrl_build_info" {
			continue
		}
		m := f.GetMetric()[0]
		if m.GetGauge().GetValue() != 1 || m.GetLabel()[0].GetValue() != BuildCommit() {
			t.Errorf("cmdctrl_build_info = %v, want 1{commit=%q}", m, BuildCommit())
		}
		return
	}
	t.Error("no cmdctrl_build_info")
}

// The listener refuses any address that is not loopback.
func TestCheckAddrRefusesNonLoopback(t *testing.T) {
	ok := []string{"127.0.0.1:9464", "127.1.2.3:9464", "[::1]:9464", "localhost:9464", "LOCALHOST:0", "127.0.0.1:0"}
	bad := []string{
		":9464",             // every interface
		"0.0.0.0:9464",      // every interface
		"[::]:9464",         // every interface
		"192.168.1.10:9464", // the LAN
		"10.0.0.5:9464",
		"cmd.labxp.io:9464", // a hostname we cannot check
		"127.0.0.1",         // no port
		"9464",
		"",
	}
	for _, a := range ok {
		if err := CheckAddr(a); err != nil {
			t.Errorf("CheckAddr(%q) = %v, want nil", a, err)
		}
	}
	for _, a := range bad {
		err := CheckAddr(a)
		if !errors.Is(err, ErrNotLoopback) {
			t.Errorf("CheckAddr(%q) = %v, want ErrNotLoopback", a, err)
		}
		if err != nil && !strings.Contains(err.Error(), AddrEnv) {
			t.Errorf("CheckAddr(%q): the error %q does not name %s", a, err, AddrEnv)
		}
	}
	if _, err := Listen("0.0.0.0:0"); !errors.Is(err, ErrNotLoopback) {
		t.Errorf("Listen(0.0.0.0:0) = %v, want ErrNotLoopback", err)
	}
}

// The listener serves the registry at GET /metrics, and only there.
func TestServerServesTheRegistry(t *testing.T) {
	ln, err := Listen("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServer(Registry, slog.New(slog.NewTextHandler(io.Discard, nil)))
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	base := "http://" + ln.Addr().String()
	resp, err := http.Get(base + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "cmdctrl_build_info") {
		t.Errorf("GET /metrics: %d, body has build info: %v", resp.StatusCode, strings.Contains(string(body), "cmdctrl_build_info"))
	}
	resp, err = http.Get(base + "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET / on the metrics listener: %d, want 404", resp.StatusCode)
	}
	if tcp := ln.Addr().(*net.TCPAddr); !tcp.IP.IsLoopback() {
		t.Errorf("bound %s", tcp)
	}
}
