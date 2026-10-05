package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// publicStack is the real public handler (newPublicHandler) over a
// minimal lobby, a memory authenticator and a hub with no authorizer,
// which accepts a ping-only socket.
func publicStack(t *testing.T) (*httptest.Server, auth.Authenticator) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	a := auth.NewMemoryAuthenticator()
	hub := ws.NewHub(log)
	lobbyHandler := lobby.Handler(lobby.Config{
		Lobby:      lobby.NewLobby(mgr),
		Auth:       a,
		AdminToken: "a-sufficiently-long-admin-token",
	})
	srv := httptest.NewServer(newPublicHandler(hub, a, nil, nil, lobbyHandler))
	t.Cleanup(func() {
		srv.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hub.Shutdown(ctx)
	})
	return srv, a
}

// series returns the value of one cmdctrl_http_requests_total series
// in metrics.Registry, and how many cmdctrl_http_request_seconds
// observations carry the route.
func series(t *testing.T, route, method, code string) (requests float64, timed uint64) {
	t.Helper()
	families, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			l := map[string]string{}
			for _, lp := range m.GetLabel() {
				l[lp.GetName()] = lp.GetValue()
			}
			switch f.GetName() {
			case "cmdctrl_http_requests_total":
				if l["route"] == route && l["method"] == method && l["code"] == code {
					requests = m.GetCounter().GetValue()
				}
			case "cmdctrl_http_request_seconds":
				if l["route"] == route {
					timed += m.GetHistogram().GetSampleCount()
				}
			}
		}
	}
	return requests, timed
}

func do(t *testing.T, srv *httptest.Server, method, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp
}

// ADR 0123 §11: the route holder labels a lobby route with its own
// pattern, not the "/" the lobby is mounted on, including through the
// rate limiter's and auth.Middleware's request copies.
func TestPublicHandlerLabelsRoutesByTheirOwnPattern(t *testing.T) {
	srv, a := publicStack(t)
	token, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New()}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		method, path, token, route string
	}{
		// Lobby, behind the per-IP rate limiter.
		{"POST", "/games/" + uuid.NewString() + "/join", "", "POST /games/{id}/join"},
		// Lobby, behind auth.Middleware.
		{"GET", "/me", token, "GET /me"},
		// The card mux, behind auth.Middleware on the public mux: the
		// card mux's own pattern, not "/cards/".
		{"GET", "/cards/" + uuid.NewString(), token, "GET /cards/{id}"},
		// Unauthenticated, the card mux is never reached: the public
		// mux's pattern.
		{"GET", "/cards/" + uuid.NewString(), "", "/cards/"},
		{"GET", "/healthz", "", "GET /healthz"},
		// The lobby's catch-all matched nothing: unmatched, not "/".
		{"GET", "/wp-login.php", "", metrics.RouteUnmatched},
	}
	classes := []string{"2xx", "3xx", "4xx", "5xx"}
	for _, c := range cases {
		before := map[string]float64{}
		for _, code := range classes {
			before[code], _ = series(t, c.route, c.method, code)
		}
		probe := do(t, srv, c.method, c.path, c.token)
		code := classes[probe.StatusCode/100-2]
		if after, _ := series(t, c.route, c.method, code); after-before[code] != 1 {
			t.Errorf("%s %s (%d): cmdctrl_http_requests_total{route=%q,method=%q,code=%q} rose by %v, want 1",
				c.method, c.path, probe.StatusCode, c.route, c.method, code, after-before[code])
		}
		if root, _ := series(t, "/", c.method, code); root != 0 {
			t.Errorf("%s %s: a request was labelled route=\"/\"", c.method, c.path)
		}
	}
}

// The WebSocket upgrades through the instrumented handler (the status
// writer passes Hijack through), is counted as 2xx, and is not timed.
func TestWebSocketIsCountedNotTimed(t *testing.T) {
	srv, _ := publicStack(t)
	before, timedBefore := series(t, wsRoute, "GET", "2xx")
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatalf("dial through the instrumented handler: %v (response %v)", err, resp)
	}
	_ = conn.Close()
	// ServeWS returns right after the upgrade, but the count lands
	// after it does: poll the monotone counter.
	deadline := time.Now().Add(5 * time.Second)
	for {
		after, timed := series(t, wsRoute, "GET", "2xx")
		if after-before >= 1 {
			if timed != timedBefore {
				t.Errorf("GET /ws was timed: %d observations, want %d", timed, timedBefore)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET /ws upgrade never counted as 2xx")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// ADR 0123 §2, §11: /metrics is not on the public mux. A request for it
// reaches the lobby's catch-all, which has no such route.
func TestMetricsIsAbsentFromThePublicMux(t *testing.T) {
	srv, _ := publicStack(t)
	for _, method := range []string{"GET", "HEAD", "POST"} {
		req, _ := http.NewRequest(method, srv.URL+"/metrics", nil)
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Errorf("%s /metrics on the public handler answered 200", method)
		}
		for _, leak := range []string{"# TYPE", "cmdctrl_", "go_goroutines"} {
			if strings.Contains(string(body), leak) {
				t.Errorf("%s /metrics on the public handler served %q", method, leak)
			}
		}
	}
	for _, r := range metrics.Routes() {
		if strings.Contains(r, "/metrics") {
			t.Errorf("a public mux registered %q", r)
		}
	}
}

// The label guard over the real routes: after driving the public
// handler, every series in the registry keeps the closed-set rule.
func TestPublicHandlerKeepsLabelsClosed(t *testing.T) {
	srv, _ := publicStack(t)
	for _, p := range []string{"/healthz", "/games", "/me", "/cards/x", "/catalog", "/roadmap", "/no/such/path", "/metrics"} {
		do(t, srv, "GET", p, "")
	}
	do(t, srv, "BREW", "/healthz", "")
	if err := metrics.CheckClosedLabels(metrics.Registry); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}

// CMDCTRL_METRICS_ADDR: unset is off, loopback is taken, and anything
// else refuses the boot naming the variable. The refusal calls
// os.Exit, so it runs in a child process.
func TestMetricsAddrConfig(t *testing.T) {
	if os.Getenv("CMDCTRL_TEST_LOADCONFIG_CHILD") == "1" {
		loadConfig(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		return
	}
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Setenv(metrics.AddrEnv, "")
	if got := loadConfig(discard).MetricsAddr; got != "" {
		t.Errorf("unset: MetricsAddr = %q, want off", got)
	}
	t.Setenv(metrics.AddrEnv, "127.0.0.1:9464")
	if got := loadConfig(discard).MetricsAddr; got != "127.0.0.1:9464" {
		t.Errorf("loopback: MetricsAddr = %q", got)
	}

	for _, bad := range []string{"0.0.0.0:9464", ":9464", "192.168.1.10:9464"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMetricsAddrConfig$")
		cmd.Env = append(os.Environ(), "CMDCTRL_TEST_LOADCONFIG_CHILD=1", metrics.AddrEnv+"="+bad)
		out, err := cmd.CombinedOutput()
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 1 {
			t.Errorf("%s=%s: boot was not refused (err %v)\n%s", metrics.AddrEnv, bad, err, out)
			continue
		}
		if !strings.Contains(string(out), metrics.AddrEnv) {
			t.Errorf("%s=%s: the refusal does not name the variable:\n%s", metrics.AddrEnv, bad, out)
		}
	}
}
