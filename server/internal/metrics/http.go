package metrics

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RouteUnmatched is the route label of a request no pattern matched:
// a 404 from a mux, a 405, or a pattern this package did not see
// registered.
const RouteUnmatched = "unmatched"

var (
	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_http_requests_total",
		Help: "HTTP requests served, by matched route pattern, method and status class.",
	}, []string{"route", "method", "code"})

	httpSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "cmdctrl_http_request_seconds",
		Help: "Time to serve an HTTP request, by matched route pattern. Long-lived routes (the WebSocket) are left out.",
		// prometheus.DefBuckets plus 30 s: a deck fetch or a cold card
		// image can take longer than 10 s, and a bucket that catches it
		// says so rather than piling it into +Inf.
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
	}, []string{"route"})
)

// methodLabels is the closed set of the method label. Anything else a
// client sends is "other": the method is client-chosen text.
var methodLabels = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace, "other",
}

func methodLabel(m string) string {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace:
		return m
	}
	return "other"
}

// codeLabels is the closed set of the code label.
var codeLabels = []string{"2xx", "3xx", "4xx", "5xx"}

// codeClass buckets a final status. Zero means the handler wrote no
// status, which net/http answers with 200, or hijacked the connection
// (a WebSocket upgrade, which writes its own 101): both succeeded, and
// both are 2xx. A status under 200 is never final, so it cannot reach
// here except as zero.
func codeClass(status int) string {
	switch {
	case status < 300:
		return "2xx"
	case status < 400:
		return "3xx"
	case status < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

// InstrumentHTTP counts and times every request through next, by the
// route pattern the innermost *ServeMux matched (see ServeMux). A
// request on a route in untimed is counted but not timed: pass the
// long-lived routes, such as "GET /ws", whose duration is a session
// length rather than a latency.
func InstrumentHTTP(next http.Handler, untimed ...string) http.Handler {
	skip := make(map[string]bool, len(untimed))
	for _, r := range untimed {
		skip[r] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		holder := &routeHolder{}
		sw := &statusWriter{ResponseWriter: w}
		r = r.WithContext(context.WithValue(r.Context(), routeKey{}, holder))

		defer func() {
			code := codeClass(sw.status)
			p := recover()
			if p != nil {
				// net/http turns a panic into a dropped connection. It
				// served nothing useful, whatever it wrote first.
				code = "5xx"
			}
			route := holder.route()
			httpRequests.WithLabelValues(route, methodLabel(r.Method), code).Inc()
			if !skip[route] {
				httpSeconds.WithLabelValues(route).Observe(time.Since(start).Seconds())
			}
			if p != nil {
				panic(p)
			}
		}()
		next.ServeHTTP(sw, r)
	})
}

// routeKey is the context key for the request's *routeHolder.
type routeKey struct{}

// routeHolder is where the muxes a request passes through record the
// pattern each matched. It is a pointer in the context, so it survives
// every r.WithContext a middleware makes on the way down: those copy
// the request, and a mux sets Pattern on its own copy, which is why the
// pattern cannot simply be read off the request at the top.
//
// A request passes through nested muxes (the public mux, then the
// lobby's or the card routes'), and each records after it dispatches,
// so the innermost records first and the outer ones after it. Each mux
// takes a depth on the way in, and a record only replaces a shallower
// one: the innermost mux's answer wins, including its "no match" (a
// lobby 404 is unmatched, not "/").
type routeHolder struct {
	mu      sync.Mutex
	entered int
	depth   int
	pattern string
}

func (h *routeHolder) enter() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entered++
	return h.entered
}

func (h *routeHolder) record(depth int, pattern string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if depth > h.depth {
		h.depth, h.pattern = depth, pattern
	}
}

func (h *routeHolder) route() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pattern == "" {
		return RouteUnmatched
	}
	return h.pattern
}

// ServeMux is an http.ServeMux that knows its own patterns. Every mux
// the server builds is one of these, so that:
//
//   - the route label is the pattern the innermost mux matched, read
//     from r.Pattern after it dispatches (Go 1.22+ sets it on the
//     request the mux was handed);
//   - the label can only take a value some mux registered. A value the
//     mux did not register (ServeMux reports a redirect target's path
//     for CONNECT, and GODEBUG=httpmuxgo121 reports nothing) is
//     unmatched, so a client cannot write its own path into a label;
//   - the label guard has the closed set to check against (Routes).
//
// Without InstrumentHTTP above it, it is a plain ServeMux.
type ServeMux struct {
	mux *http.ServeMux

	mu       sync.RWMutex
	patterns map[string]bool
}

// NewServeMux returns an empty *ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{mux: http.NewServeMux(), patterns: map[string]bool{}}
}

// Handle registers handler for pattern, as http.ServeMux.Handle does
// (and panics as it does on a bad or conflicting pattern).
func (m *ServeMux) Handle(pattern string, handler http.Handler) {
	m.mux.Handle(pattern, handler)
	m.mu.Lock()
	m.patterns[pattern] = true
	m.mu.Unlock()
	routes.add(pattern)
}

// HandleFunc registers f for pattern, as http.ServeMux.HandleFunc does.
func (m *ServeMux) HandleFunc(pattern string, f func(http.ResponseWriter, *http.Request)) {
	if f == nil {
		panic("http: nil handler")
	}
	m.Handle(pattern, http.HandlerFunc(f))
}

// ServeHTTP dispatches as http.ServeMux does, then records the matched
// pattern in the request's route holder, if it has one.
func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	holder, _ := r.Context().Value(routeKey{}).(*routeHolder)
	if holder == nil {
		m.mux.ServeHTTP(w, r)
		return
	}
	depth := holder.enter()
	// Deferred so a handler that panics still has its route: the mux
	// sets r.Pattern before it calls the handler.
	defer func() { holder.record(depth, m.own(r.Pattern)) }()
	m.mux.ServeHTTP(w, r)
}

// own returns pattern if this mux registered it, and "" otherwise.
func (m *ServeMux) own(pattern string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.patterns[pattern] {
		return pattern
	}
	return ""
}

// routes is every pattern registered on any *ServeMux in this process:
// the closed set of the route label, less RouteUnmatched.
var routes = &routeSet{set: map[string]bool{}}

type routeSet struct {
	mu  sync.RWMutex
	set map[string]bool
}

func (s *routeSet) add(p string) {
	s.mu.Lock()
	s.set[p] = true
	s.mu.Unlock()
}

func (s *routeSet) has(p string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set[p]
}

// Routes returns every pattern registered on a *ServeMux in this
// process, sorted.
func Routes() []string {
	routes.mu.RLock()
	defer routes.mu.RUnlock()
	out := make([]string, 0, len(routes.set))
	for p := range routes.set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// statusWriter records the first final status a handler writes. It
// passes Flush, Hijack and ReadFrom through, because the WebSocket
// upgrade needs Hijack (gorilla/websocket asserts http.Hijacker) and
// http.ServeFile's sendfile path needs ReadFrom; Unwrap serves
// http.ResponseController.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 && code >= 200 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) ReadFrom(src io.Reader) (int64, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	return io.Copy(writerOnly{w.ResponseWriter}, src)
}

func (w *statusWriter) Flush() {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("metrics: the underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// writerOnly hides every method but Write, so io.Copy cannot find
// ReadFrom on it and recurse.
type writerOnly struct{ io.Writer }
